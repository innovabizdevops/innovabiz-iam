# INNOVABIZ-OPS-AM01 - AlertManager Operational Runbook

## 🚨 Visão Geral

**Módulo:** INNOVABIZ Observability Framework  
**Componente:** AlertManager  
**Versão:** 0.26.0  
**Equipe Responsável:** Observability Team  
**Contatos Primários:** observability@innovabiz.com, #observability-support  
**Repositório:** `CoreModules/IAM/docs/observability`  

## 📋 Informações Gerais

O AlertManager é um componente crítico da pilha de observabilidade INNOVABIZ, responsável pelo gerenciamento, agrupamento, roteamento e entrega de alertas gerados pelo Prometheus. Este runbook fornece orientações para operações, troubleshooting e manutenção do AlertManager implantado nos clusters Kubernetes INNOVABIZ.

## 🔐 Multi-Contexto e Compliance

O AlertManager da INNOVABIZ opera em conformidade com:

- **Multi-Tenant:** Isolamento completo por tenant com roteamento específico
- **Multi-Regional:** Configuração adaptada por região (BR, US, EU, AO)
- **Multi-Ambiente:** Separação entre produção, homologação e desenvolvimento
- **Multi-Moeda:** Contextos financeiros por moeda (BRL, USD, EUR, AOA)
- **Multi-Idioma:** Templates de notificação localizados (pt-BR, pt-AO, en-US, en-GB)
- **Compliance:** PCI DSS 4.0, GDPR/LGPD, ISO 27001, NIST CSF, SOX, Basel III

## 📊 Arquitetura e Componentes

### Componentes Principais

- **AlertManager UI:** Interface web para visualização e gerenciamento de alertas
- **Cluster Mesh:** Sincronização entre múltiplas instâncias do AlertManager
- **Silence API:** API para criação e gerenciamento de silenciamentos
- **Notification Pipeline:** Sistema de processamento e entrega de notificações
- **Canais Configurados:** Email, Slack, PagerDuty, OpsGenie, Webhooks

### Interdependências

1. **Prometheus:** Origem primária dos alertas
2. **Grafana:** Visualização alternativa dos alertas ativos
3. **KrakenD:** API Gateway para exposição segura da API
4. **IAM Service:** Autenticação e autorização
5. **Redis:** Cache para estado compartilhado (opcional)
6. **Serviços SMTP:** Para notificações por email

## 🔍 Procedimentos Operacionais

### 🟢 Verificação de Status

```bash
# Verificar estado dos pods do AlertManager
kubectl get pods -n observability-${TENANT_ID}-${REGION} -l app=alertmanager

# Verificar logs em tempo real
kubectl logs -f -n observability-${TENANT_ID}-${REGION} -l app=alertmanager

# Acessar a interface web (via port-forward para testes)
kubectl port-forward -n observability-${TENANT_ID}-${REGION} svc/alertmanager 9093:9093
# Acessar: http://localhost:9093
```

### 🟢 Verificações de Saúde

```bash
# Status do componente
curl -sSL http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/-/healthy

# Verificar métricas expostas
curl -sSL http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/metrics | grep alertmanager

# Status do cluster (informações sobre membros do cluster mesh)
curl -sSL http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/-/ready
```

### 🔄 Manutenção de Rotina

#### Atualizações de Versão

1. Consulte o [Changelog oficial](https://github.com/prometheus/alertmanager/releases) para impactos
2. Atualize a versão no manifesto YAML ou Helm Chart
3. Realize o deploy em ambientes de menor criticidade primeiro
4. Monitore métricas de performance e integridade após a atualização
5. Execute smoke tests para garantir entrega de notificações

```bash
# Atualizar a versão via kubectl
kubectl -n observability-${TENANT_ID}-${REGION} set image deployment/alertmanager \
  alertmanager=quay.io/prometheus/alertmanager:v0.26.0
```

#### Backups de Configuração

```bash
# Backup das configurações e silenciamentos
kubectl -n observability-${TENANT_ID}-${REGION} get configmap alertmanager-config -o yaml > alertmanager-config-backup-$(date +%Y%m%d).yaml

# Backup do persistent volume (se aplicável)
kubectl -n observability-${TENANT_ID}-${REGION} get pvc alertmanager-storage -o yaml > alertmanager-pvc-backup-$(date +%Y%m%d).yaml
```

## 🔎 Troubleshooting

### Problema: Alertas Não Estão Sendo Enviados

#### Diagnóstico

1. Verificar se os alertas estão chegando ao AlertManager:
   ```bash
   curl -s http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/api/v1/alerts | jq
   ```

2. Verificar logs para erros de entrega:
   ```bash
   kubectl logs -f -n observability-${TENANT_ID}-${REGION} -l app=alertmanager | grep -E "error|failed"
   ```

3. Verificar a configuração de roteamento:
   ```bash
   kubectl -n observability-${TENANT_ID}-${REGION} get configmap alertmanager-config -o yaml
   ```

#### Resolução

1. **Problemas de conectividade SMTP:**
   - Verificar credenciais SMTP no secret
   - Testar conectividade ao servidor SMTP
   - Revisar logs para erros específicos de SMTP

2. **Problemas com Slack/API externa:**
   - Validar token/URL no secret
   - Verificar conectividade com a API externa
   - Revisar políticas de rede para egress

3. **Problemas de roteamento:**
   - Revisar labels usados nos alertas
   - Verificar regras de match/mismatch
   - Verificar se receivers estão corretamente configurados

### Problema: Alta Carga de Alertas ou Performance Degradada

#### Diagnóstico

1. Verificar métricas de performance:
   ```bash
   curl -s http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/metrics | grep -E "(alertmanager_alerts|alertmanager_notifications|process_cpu|process_resident)"
   ```

2. Analisar número de alertas por tenant/região:
   ```bash
   curl -s http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/api/v1/alerts | jq '.[] | .labels.tenant_id' | sort | uniq -c
   ```

#### Resolução

1. **Alta carga de alertas:**
   - Revisar limites de agrupamento
   - Aumentar recursos do deployment
   - Escalar horizontalmente (aumentar réplicas)

2. **Problemas de performance:**
   - Verificar CPU/memória e ajustar limites
   - Otimizar regras de agrupamento para reduzir processamento
   - Avaliar configurações de throttling para receivers

### Problema: Sincronização entre Instâncias do AlertManager

#### Diagnóstico

1. Verificar status do cluster:
   ```bash
   curl -s http://alertmanager.${TENANT_ID}.${REGION}.innovabiz.io/api/v1/status | jq '.cluster'
   ```

2. Verificar conectividade entre pods:
   ```bash
   kubectl -n observability-${TENANT_ID}-${REGION} exec -it alertmanager-0 -- wget -qO- alertmanager-1.alertmanager:9094/metrics
   ```

#### Resolução

1. **Problemas de sincronização:**
   - Verificar NetworkPolicy permitindo tráfego entre pods
   - Garantir que ports 9094 estão abertos para cluster mesh
   - Validar que todos os pods têm a mesma configuração de cluster

## 🔒 Segurança e Compliance

### Controles de Segurança Implementados

- TLS para toda comunicação externa (via Ingress)
- Autenticação básica na interface web
- Execução como usuário não-root (65534)
- Filesystem read-only
- NetworkPolicy restritiva
- Secrets para credenciais sensíveis
- Audit logging ativado
- Headers de segurança HTTP (Strict-Transport-Security, X-Content-Type-Options)
- Prevenção de CSRF e XSS

### Verificações de Compliance

```bash
# Verificar permissões RBAC
kubectl -n observability-${TENANT_ID}-${REGION} get role alertmanager-role -o yaml

# Verificar SecurityContext
kubectl -n observability-${TENANT_ID}-${REGION} get pods -l app=alertmanager -o jsonpath='{.items[0].spec.containers[0].securityContext}'

# Verificar configurações de NetworkPolicy
kubectl -n observability-${TENANT_ID}-${REGION} get networkpolicy alertmanager-network-policy -o yaml
```

## 📈 Monitoramento e Métricas

### Métricas Críticas

| Métrica | Descrição | Thresholds |
|---------|-----------|------------|
| `alertmanager_alerts` | Número total de alertas | Alerta: >1000 |
| `alertmanager_alerts_received_total` | Total de alertas recebidos | Tendência: >30% do normal |
| `alertmanager_notifications_failed_total` | Notificações com falha | Alerta: >5% |
| `alertmanager_notification_latency_seconds` | Latência de notificação | Alerta: >10s |
| `alertmanager_cluster_members` | Membros do cluster | Alerta: <configurado |
| `process_cpu_seconds_total` | Utilização de CPU | Alerta: >80% |
| `process_resident_memory_bytes` | Utilização de memória | Alerta: >80% |
| `alertmanager_silences` | Número de silenciamentos ativos | Info: monitorar tendência |
| `alertmanager_nflog_gc_duration_seconds` | Tempo de garbage collection | Alerta: >5s |

### Dashboards Recomendados

- **AlertManager Overview:** Visão geral de saúde e performance
- **AlertManager por Tenant:** Métricas separadas por tenant
- **AlertManager por Região:** Métricas separadas por região
- **Alert Delivery SLOs:** Métricas de SLO para entrega de alertas
- **Alertas Multi-Dimensionais:** Visualização por tenant/região/ambiente
- **Alert Heatmap:** Distribuição temporal de alertas

## 📚 Melhores Práticas

### Design de Alertas

- **Severidade Consistente:** Usar labels `severity` padronizados (critical, warning, info)
- **Contexto Multi-dimensional:** Incluir sempre `tenant_id`, `region_id`, `environment`, `component`
- **Acionabilidade:** Cada alerta deve ter uma ação clara a ser tomada
- **Documentação:** Incluir links para runbooks ou documentação em `annotations.runbook_url`
- **Correlação:** Usar labels que permitam correlacionar alertas relacionados
- **Precisão:** Evitar falsos positivos com thresholds bem calibrados
- **Priorização:** Usar `priority` label para indicar urgência (P1-P5)

### Agrupamento Eficiente

- Agrupar por `tenant_id`, `severity`, `alertname` e `region_id`
- Estabelecer intervalos de grupo adequados por severidade
- Evitar "alert storms" com regras de inibição apropriadas
- Utilizar silêncios com data de expiração e justificativa
- Implementar routes específicas para diferentes tipos de alerta
- Configurar timeouts e repetições adequadas por severidade
- Utilizar continue: true para encaminhamento multi-canal quando necessário

### Notificações Efetivas

- Personalizar templates por canal de notificação
- Incluir links para Grafana e sistemas de ticketing
- Formatação clara com indicadores visuais de severidade
- Enviar para canais apropriados com base no horário e escalation path
- Incluir métricas relevantes e contexto do problema
- Localizar mensagens conforme idioma regional (pt-BR, pt-AO, en-US, etc)
- Adicionar informações de SLA baseadas em severidade e tenant

## 📆 Manutenção e Ciclo de Vida

### Verificações Periódicas

| Frequência | Atividade | Responsável |
|------------|-----------|-------------|
| Diário | Validar recebimento e processamento de alertas | SRE On-call |
| Semanal | Revisar métricas de desempenho | DevOps |
| Semanal | Revisar silenciamentos expirados | DevOps |
| Mensal | Testar recuperação de backup | SRE Team |
| Mensal | Revisar alertas mais frequentes para redução de ruído | Observability Team |
| Trimestral | Revisar configurações e rotas | Observability Team |
| Trimestral | Teste de carga do sistema de alertas | Performance Team |
| Anual | Avaliar atualizações de versão | Platform Team |
| Anual | Auditoria completa de segurança e compliance | Security Team |

### Melhorias Contínuas

- Revisar alertas mais frequentes para redução de ruído
- Otimizar templates de notificação para maior clareza
- Refinar regras de agrupamento para melhor organização
- Calibrar thresholds com base em dados históricos
- Analisar tempos de resposta e resolução para aprimoramento
- Implementar automações para alertas repetitivos
- Desenvolver machine learning para detecção de anomalias
- Integrar com sistemas de análise pós-incidente

## 🔗 Integração com Outros Componentes

### Prometheus

```yaml
# Exemplo de configuração no Prometheus
alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager.${TENANT_ID}.${REGION}.innovabiz.io
    api_version: v2
    timeout: 10s
    scheme: https
    tls_config:
      ca_file: /etc/prometheus/certs/ca.crt
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_tenant_id]
        target_label: tenant_id
      - source_labels: [__meta_kubernetes_pod_label_region_id]
        target_label: region_id
```

### Grafana

- Configurar AlertManager como data source em Grafana
- Utilizar integração para Unified Alerting quando aplicável
- Configurar dashboards específicos para visualização de alertas
- Implementar anotações em gráficos para visualizar eventos de alerta

### OpsGenie/PagerDuty

- Configurar integração bidirecional para sincronização de status
- Mapeamento de severidades entre sistemas
- Escalation policies alinhadas à estrutura organizacional
- Tags e atributos para preservar contexto multi-dimensional

### ITSM/ServiceNow

- Criação automática de tickets baseada em alertas
- Campos personalizados para contexto multi-dimensional
- Regras de prioridade e categorização alinhadas
- Fechamento automático baseado em resolução de alerta

## 📝 Documentação Relacionada

- [INNOVABIZ Alerting Framework](../ALERTING_FRAMEWORK.md)
- [INNOVABIZ Prometheus Template](./KUBERNETES_YAML/PROMETHEUS_TEMPLATE.md)
- [INNOVABIZ AlertManager Template](./KUBERNETES_YAML/ALERTMANAGER_TEMPLATE.md)
- [INNOVABIZ Observability Policy](../../../policies/OBSERVABILITY_POLICY.md)
- [Documentação Oficial AlertManager](https://prometheus.io/docs/alerting/latest/alertmanager/)

## 📞 Suporte e Escalonamento

| Nível | Contato | SLA |
|-------|---------|-----|
| L1 | SRE On-call: sre-oncall@innovabiz.com | 15min |
| L2 | DevOps Team: devops@innovabiz.com | 30min |
| L3 | Observability Team: observability@innovabiz.com | 60min |

**Em caso de emergência:** +55 11 95555-1234

**Canais de comunicação:**
- Slack: #observability-support, #sre-alerts
- Email: observability@innovabiz.com
- Teams: Observability Team
- Horário comercial: 08:00-18:00 (todas as regiões suportadas)

## 📅 Histórico de Revisões

| Data | Versão | Autor | Descrição |
|------|--------|-------|-----------|
| 31/07/2025 | 1.0.0 | Eduardo Jeremias | Versão inicial |