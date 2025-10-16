# INNOVABIZ Kubernetes Cluster Overview Dashboard

## Visão Geral

O **INNOVABIZ Kubernetes Cluster Overview Dashboard** é uma solução de observabilidade avançada que oferece uma visão completa e em tempo real do estado e desempenho de clusters Kubernetes em múltiplos contextos (tenant, região e ambiente). Este dashboard foi projetado para atender às necessidades de monitoramento de infraestrutura de equipes DevOps, SRE e operações na plataforma INNOVABIZ, proporcionando visibilidade total sobre a saúde e performance do cluster Kubernetes.

![Dashboard Preview](../../../images/kubernetes_dashboard_preview.png)

## Características Principais

- **Monitoramento Multi-Contexto**: Filtragem por tenant ID, região e ambiente (produção, staging, desenvolvimento)
- **Métricas de Estado do Cluster**: Visão geral do número total de nodes, namespaces e pods
- **Estado dos Nodes**: Monitoramento de disponibilidade e estado dos nodes (Ready, NotReady)
- **Uso de Recursos**: Visualização de CPU, memória, disco e rede por node
- **Estado de Workloads**: Contagens e disponibilidade de Deployments, StatefulSets, DaemonSets e Jobs
- **Estado dos Pods**: Monitoramento de pods em execução, com falha, pendentes e reinícios de containers
- **Métricas de Rede**: Tráfego de entrada e saída por node
- **Visualização de Tendências**: Gráficos de séries temporais para análise de tendências
- **Alertas Visuais**: Codificação por cores para identificação rápida de problemas
- **Atualização Automática**: Atualização periódica a cada 30 segundos

## Estrutura do Dashboard

O dashboard está organizado em três seções principais:

### 1. Cluster Status

Esta seção fornece uma visão geral do estado atual do cluster, incluindo:

- **Total de Nodes**: Contagem total de nodes no cluster
- **Namespaces**: Número total de namespaces ativos
- **Nodes Ready**: Percentual e contagem de nodes em estado Ready
- **Nodes Not Ready**: Contagem de nodes em estado NotReady
- **Pods Running**: Contagem de pods em execução
- **Pods Failed/Pending**: Contagem de pods em estado falho ou pendente
- **Container Restarts**: Contagem de reinícios de containers nas últimas 24 horas

### 2. Cluster Resources

Esta seção apresenta o uso de recursos por node no cluster:

- **CPU Utilization by Node**: Gráfico de série temporal do uso de CPU por node
- **Memory Utilization by Node**: Gráfico de série temporal do uso de memória por node
- **Disk Utilization by Node**: Gráfico de série temporal do uso de disco por node
- **Network Traffic by Node**: Gráfico de série temporal do tráfego de rede por node (recebido/transmitido)

### 3. Workloads

Esta seção fornece insights sobre o estado das cargas de trabalho no cluster:

- **Workload Counts**: Contagem de Deployments, StatefulSets, DaemonSets e Jobs
- **Deployment Readiness**: Percentual de réplicas disponíveis vs. desejadas em Deployments
- **Pods Running Rate**: Percentual de pods em estado Running vs. total

## Variáveis e Filtragem Multi-Contexto

O dashboard suporta filtragem dinâmica através das seguintes variáveis:

- **DS_PROMETHEUS**: Fonte de dados Prometheus
- **tenant_id**: Identificador do tenant para isolamento multi-tenant
- **region_id**: Identificador de região geográfica
- **environment**: Ambiente (production, staging, development)

Estas variáveis permitem o monitoramento específico por contexto, garantindo alinhamento com a arquitetura multi-contexto da plataforma INNOVABIZ.

## Implementação e Uso

### Pré-requisitos

Para implementar e utilizar este dashboard, é necessário:

1. **Grafana 9.0+** configurado e acessível
2. **Prometheus 2.40+** coletando métricas de Kubernetes
3. **kube-state-metrics** instalado no cluster Kubernetes
4. **node-exporter** instalado nos nodes do cluster
5. **Métricas Kubernetes** etiquetadas com `tenant_id`, `region_id` e `environment`

### Procedimento de Importação

1. Acesse o Grafana da sua instância INNOVABIZ
2. Navegue até Dashboards > Import
3. Faça upload do arquivo JSON `KUBERNETES_CLUSTER_OVERVIEW_DASHBOARD.json`
4. Selecione a fonte de dados Prometheus apropriada
5. Clique em "Import" para concluir

### Etiquetagem de Métricas

Para garantir o funcionamento adequado da filtragem multi-contexto, as métricas devem ser etiquetadas corretamente:

```yaml
prometheus:
  serviceMonitor:
    relabelings:
      - sourceLabels: [__meta_kubernetes_namespace]
        targetLabel: tenant_id
        regex: '(.+)-.*'
        replacement: '$1'
      - targetLabel: region_id
        replacement: 'br-sp'  # Exemplo: região São Paulo
      - targetLabel: environment
        replacement: 'production'  # Ambiente: production, staging, development
```

## Casos de Uso Operacional

### Monitoramento Diário

- **Equipes SRE/DevOps**: Verificação rápida da saúde geral do cluster Kubernetes
- **Plantão**: Identificação de problemas emergentes antes que afetem usuários
- **Operações**: Validação de disponibilidade de serviços após implantações

### Análise de Problemas

- **Investigação de Falhas**: Correlação entre eventos de cluster e degradação de serviço
- **Capacidade de Recursos**: Análise de tendências de uso para planejamento de capacidade
- **Troubleshooting de Pods**: Identificação de pods problemáticos ou com reinícios frequentes

### Governança e Compliance

- **Auditoria**: Verificação de status operacional para relatórios de conformidade
- **SLOs/SLIs**: Monitoramento de indicadores de nível de serviço
- **Validação Multi-Tenant**: Verificação de isolamento apropriado entre tenants

## Manutenção e Governança

### Atualização e Versionamento

O dashboard deve seguir as práticas de versionamento e atualização estabelecidas:

- Versionamento semântico (Major.Minor.Patch)
- Documentação de alterações em cada versão
- Testes em ambiente de homologação antes da promoção para produção
- Armazenamento do código JSON em repositório Git

### Revisão e Aprovação

Qualquer alteração ao dashboard deve seguir o processo de governança:

1. Solicitação de alteração documentada
2. Implementação em ambiente de desenvolvimento
3. Revisão por equipe técnica
4. Aprovação por responsáveis de plataforma
5. Implementação em produção

### Ciclo de Vida

- **Revisão Trimestral**: Validação de relevância e precisão das métricas
- **Atualização Anual**: Revisão completa e alinhamento com evolução da plataforma
- **Documentação**: Manutenção de documentação atualizada sobre o dashboard

## Integração com Plataforma INNOVABIZ

Este dashboard é parte integrante do ecossistema de observabilidade da plataforma INNOVABIZ e está alinhado com:

### Frameworks e Normas

- **ISO/IEC 27001**: Segurança da informação
- **NIST CSF**: Framework de Segurança Cibernética
- **PCI DSS 4.0**: Requisitos 10.2, 10.3 (monitoramento e logging)
- **COBIT**: Monitoramento, avaliação e análise (MEA)

### Integração com Outros Componentes

- **AlertManager**: Correlação com alertas gerados
- **Logs Centralizados**: Correlação com eventos de log
- **API Gateway**: Monitoramento de impacto em APIs expostas
- **Sistema ITSM**: Integração com tickets de incidentes

## Recomendações de Uso

### Melhores Práticas

- **Verificação Regular**: Estabelecer rotinas de verificação periódica (início/fim de expediente)
- **Alertas Complementares**: Configurar alertas para métricas críticas observadas
- **Correlação**: Usar IDs de correlação para relacionar eventos entre dashboards
- **Documentação de Incidentes**: Referenciar dashboards em relatórios de incidentes

### Treinamento

- Incluir este dashboard em treinamentos para novos membros da equipe
- Documentar casos de uso específicos e interpretações de padrões observados
- Criar runbooks associados para resposta a cenários comuns

## Considerações de Segurança

- O acesso a este dashboard deve ser controlado via IAM da plataforma INNOVABIZ
- Implementar restrição de visualização por tenant_id conforme permissões do usuário
- Garantir que dados sensíveis não sejam expostos nas métricas visualizadas
- Manter registros de auditoria de acesso ao dashboard

## Contatos e Suporte

Para questões relacionadas a este dashboard:

- **Responsável Técnico**: Equipe DevOps INNOVABIZ
- **E-mail**: devops@innovabiz.com
- **Canal Slack**: #innovabiz-observability
- **Documentação Adicional**: [Wiki Interna - Observabilidade](https://wiki.innovabiz.com/observability/kubernetes)

## Histórico de Versões

| Versão | Data       | Autor          | Descrição das Alterações                          |
|--------|------------|----------------|---------------------------------------------------|
| 1.0.0  | 2025-07-29 | Eduardo Jeremias | Versão inicial do dashboard                      |

---

*Este documento é parte integrante da documentação técnica da plataforma INNOVABIZ e deve ser mantido atualizado conforme evolução do sistema de observabilidade.*