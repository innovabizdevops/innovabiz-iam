# Guia de Operações - Módulo Bureau de Crédito

![INNOVABIZ](../../../assets/images/logo.png)

**Versão:** 1.0.0  
**Data:** 2025-08-06  
**Classificação:** Privado  
**Autor:** Equipe de Operações INNOVABIZ

## Índice

1. [Introdução](#1-introdução)
2. [Visão Geral Operacional](#2-visão-geral-operacional)
3. [Infraestrutura e Implantação](#3-infraestrutura-e-implantação)
4. [Monitoramento e Observabilidade](#4-monitoramento-e-observabilidade)
5. [Resolução de Problemas](#5-resolução-de-problemas)
6. [Procedimentos de Manutenção](#6-procedimentos-de-manutenção)
7. [Procedimentos de Backup e Recuperação](#7-procedimentos-de-backup-e-recuperação)
8. [Políticas de Segurança Operacional](#8-políticas-de-segurança-operacional)
9. [Conformidade por Mercado](#9-conformidade-por-mercado)
10. [Gestão de Configuração](#10-gestão-de-configuração)
11. [Referências](#11-referências)

## 1. Introdução

Este guia de operações fornece instruções detalhadas para a administração, manutenção, monitoramento e resolução de problemas do módulo Bureau de Crédito na plataforma INNOVABIZ. O documento foi desenvolvido para equipes de operações, DevSecOps e suporte que gerenciam o ambiente produtivo e garantem a disponibilidade, performance, segurança e conformidade regulatória do módulo.

### 1.1 Objetivos do Módulo

O Bureau de Crédito é um componente crítico da plataforma INNOVABIZ que:
- Fornece consultas de crédito em diferentes níveis de profundidade
- Executa validações de compliance específicas por mercado
- Gerencia consentimentos para consultas de dados financeiros
- Integra-se com outros módulos core (IAM, Payment Gateway, Risk Management, Mobile Money, Marketplace)
- Mantém registros de auditoria para todas as consultas e ações

### 1.2 Contexto Operacional

O módulo opera em um ambiente:
- **Multi-mercado**: Suporta regras específicas para Angola, Brasil, União Europeia, EUA e mercado global
- **Multi-tenant**: Permite configurações por tipo de tenant e necessidades específicas
- **Multi-contexto**: Adapta comportamento baseado no contexto de mercado e aplicação
- **Altamente observável**: Gera telemetria completa para auditoria, compliance e operações

### 1.3 Requisitos Operacionais

| Requisito | Descrição | Meta |
|-----------|-----------|------|
| Disponibilidade | Percentual de tempo em que o serviço está acessível | 99.95% |
| MTTR | Tempo médio para recuperação após falha | < 15 minutos |
| Latência P95 | Tempo de resposta para 95% das consultas | < 500ms |
| Taxa de erro | Percentual de consultas com erro técnico | < 0.1% |
| RTO | Objetivo de tempo de recuperação | < 1 hora |
| RPO | Objetivo de ponto de recuperação | < 5 minutos |
| Alertas | Tempo para notificação de incidentes | < 1 minuto |

## 2. Visão Geral Operacional

### 2.1 Arquitetura Operacional

O Bureau de Crédito está implementado como um serviço em contêineres, implantado em um cluster Kubernetes gerenciado. A arquitetura operacional inclui:

```
                                       ┌────────────────┐
                                       │   API Gateway  │
                                       │    KrakenD     │
                                       └───────┬────────┘
                                               │
                ┌──────────────────────────────┴───────────────────────────┐
                │                                                          │
        ┌───────┴───────┐                                        ┌────────┴──────────┐
        │  Bureau de    │                                        │                   │
        │   Crédito     │◄─────────────┐             ┌──────────►│  IAM Service      │
        │  Service      │              │             │           │                   │
        └───┬───────────┘              │             │           └───────────────────┘
            │                          │             │
            │                 ┌────────┴─────────┐   │
        ┌───┴───────────┐     │                  │   │        ┌─────────────────────┐
        │ Prometheus/   │     │  OpenTelemetry   │   │        │                     │
        │ Grafana       │     │  Collector       │◄──┴────────┤  Logs/Audit Storage │
        │               │     │                  │            │                     │
        └───────────────┘     └──────────────────┘            └─────────────────────┘
```

### 2.2 Componentes Operacionais

| Componente | Descrição | Responsabilidade |
|------------|-----------|------------------|
| Bureau de Crédito Service | Serviço principal do módulo | Processamento de consultas de crédito, aplicação de regras de compliance |
| API Gateway (KrakenD) | Gateway de API | Exposição segura de APIs, autenticação, autorização e limitação de taxa |
| IAM Service | Serviço de identidade | Autenticação, autorização e gestão de consentimentos |
| OpenTelemetry Collector | Coletor de telemetria | Coleta centralizada de traces, métricas e logs |
| Prometheus/Grafana | Stack de monitoramento | Armazenamento de métricas, visualização e alertas |
| Logs/Audit Storage | Armazenamento de logs e auditoria | Armazenamento centralizado de logs e eventos de auditoria |

### 2.3 Fluxos Operacionais

#### 2.3.1 Fluxo de Consulta de Crédito

1. A requisição é recebida pelo KrakenD API Gateway
2. O Gateway valida o token JWT e escopos de acesso
3. A requisição é encaminhada para o Bureau de Crédito Service
4. O serviço valida autenticação, autorização, consentimento e compliance
5. São aplicados limites de consulta diária
6. A consulta é processada e registrada para auditoria
7. Notificações regulatórias são enviadas quando necessário
8. O resultado é retornado ao cliente

#### 2.3.2 Fluxo de Telemetria

1. Cada operação gera spans OpenTelemetry
2. Eventos de auditoria são registrados para todas as consultas
3. Eventos de segurança são gerados para tentativas inválidas
4. Métricas de negócio e operacionais são coletadas
5. O OpenTelemetry Collector agrega e encaminha a telemetria
6. Prometheus armazena métricas para alertas e dashboards
7. Grafana apresenta dashboards operacionais e de negócios

### 2.4 Modelos de Suporte

| Nível | Tempo de Resposta | Horário de Cobertura | Equipe Responsável |
|-------|-------------------|----------------------|-------------------|
| L1 | 15 minutos | 24x7 | NOC / Operações |
| L2 | 1 hora | 24x7 | DevSecOps |
| L3 | 4 horas | Horário comercial | Desenvolvimento |

## 3. Infraestrutura e Implantação

### 3.1 Requisitos de Infraestrutura

| Recurso | Mínimo | Recomendado | Observações |
|---------|--------|-------------|-------------|
| CPU | 2 cores | 4 cores | Por réplica |
| Memória | 1 GB | 2 GB | Por réplica |
| Armazenamento | 10 GB | 20 GB | Para cache local |
| Réplicas | 3 | 5+ | Distribuídas em AZs |
| Rede | 100 Mbps | 1 Gbps | Baixa latência |

### 3.2 Manifests Kubernetes

#### 3.2.1 Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bureau-credito
  namespace: innovabiz
  labels:
    app: bureau-credito
    module: core
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bureau-credito
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: bureau-credito
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - bureau-credito
              topologyKey: kubernetes.io/hostname
      containers:
      - name: bureau-credito
        image: innovabiz/bureau-credito:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: ENVIRONMENT
          valueFrom:
            configMapKeyRef:
              name: bureau-credito-config
              key: ENVIRONMENT
        - name: MARKET
          valueFrom:
            configMapKeyRef:
              name: bureau-credito-config
              key: MARKET
        - name: TENANT_TYPE
          valueFrom:
            configMapKeyRef:
              name: bureau-credito-config
              key: TENANT_TYPE
        - name: SERVICE_VERSION
          valueFrom:
            configMapKeyRef:
              name: bureau-credito-config
              key: SERVICE_VERSION
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: bureau-credito-config
              key: LOG_LEVEL
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: bureau-credito-config
              key: OTEL_EXPORTER_OTLP_ENDPOINT
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
          requests:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        volumeMounts:
        - name: config-volume
          mountPath: /app/config
        - name: certs-volume
          mountPath: /app/certs
      volumes:
      - name: config-volume
        configMap:
          name: bureau-credito-config
      - name: certs-volume
        secret:
          secretName: bureau-credito-certs
```

#### 3.2.2 Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: bureau-credito
  namespace: innovabiz
  labels:
    app: bureau-credito
    module: core
spec:
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: bureau-credito
  type: ClusterIP
```

#### 3.2.3 ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: bureau-credito-config
  namespace: innovabiz
data:
  ENVIRONMENT: "production"
  MARKET: "global"
  TENANT_TYPE: "default"
  SERVICE_VERSION: "1.0.0"
  LOG_LEVEL: "info"
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://otel-collector.observability:4317"
  REDIS_HOST: "redis-master.cache"
  REDIS_PORT: "6379"
  DAILY_QUERY_LIMIT_DEFAULT: "100"
  DAILY_QUERY_LIMIT_ANGOLA: "50"
  DAILY_QUERY_LIMIT_BRAZIL: "200"
  DAILY_QUERY_LIMIT_EU: "100"
  DAILY_QUERY_LIMIT_USA: "150"
  NOTIFICATION_REQUIRED_ANGOLA: "true"
  NOTIFICATION_REQUIRED_BRAZIL: "false"
  NOTIFICATION_REQUIRED_EU: "true"
  NOTIFICATION_REQUIRED_USA: "false"
```

#### 3.2.4 HPA (Horizontal Pod Autoscaler)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: bureau-credito-hpa
  namespace: innovabiz
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: bureau-credito
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Pods
        value: 1
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 4
        periodSeconds: 15
      selectPolicy: Max
```

### 3.3 Processo de Implantação

#### 3.3.1 Estratégia de Implantação

O módulo Bureau de Crédito utiliza implantação Blue/Green através do ArgoCD com as seguintes fases:

1. **Build & Test**: Compilação e testes automatizados (unitários, integração, performance)
2. **Publicação de Imagem**: Publicação da imagem Docker no registry corporativo
3. **Deploy em Homologação**: Implantação automatizada em ambiente de homologação
4. **Testes de Aceitação**: Testes automatizados de aceitação
5. **Aprovação**: Aprovação manual ou automatizada baseada em métricas
6. **Deploy em Produção**: Implantação em produção usando Blue/Green
7. **Validação**: Validação de integridade e performance
8. **Finalização**: Promoção da nova versão e remoção da antiga

#### 3.3.2 Comandos de Implantação

Para implantação manual (somente em emergências):

```bash
# Aplicar ConfigMap
kubectl apply -f bureau-credito-configmap.yaml -n innovabiz

# Aplicar Deployment
kubectl apply -f bureau-credito-deployment.yaml -n innovabiz

# Monitorar rollout
kubectl rollout status deployment/bureau-credito -n innovabiz

# Em caso de problemas, rollback
kubectl rollout undo deployment/bureau-credito -n innovabiz
```

Para implantação via ArgoCD:

```bash
# Sincronizar aplicação
argocd app sync bureau-credito

# Verificar status da sincronização
argocd app get bureau-credito

# Promover versão (Blue/Green)
argocd app actions run bureau-credito promote --kind Rollout
```## 4. Monitoramento e Observabilidade

### 4.1 Estratégia de Observabilidade

O Bureau de Crédito implementa uma estratégia de observabilidade completa baseada nos quatro pilares:

1. **Traces**: Rastreamento distribuído de todas as transações e chamadas internas/externas
2. **Métricas**: Medições quantitativas de desempenho, utilização e negócio
3. **Logs**: Registros estruturados de eventos e ações do sistema
4. **Eventos**: Auditoria e segurança detalhados para conformidade regulatória

Esta estratégia permite:
- Visibilidade end-to-end das consultas de crédito
- Identificação rápida de problemas e gargalos
- Análise de causa raiz de incidentes
- Verificação de compliance em tempo real
- Detecção de padrões anômalos e potenciais fraudes

### 4.2 Métricas Chave

| Categoria | Métrica | Descrição | Alertas |
|-----------|---------|-----------|---------|
| **Disponibilidade** | `bureau_credito_health` | Estado de saúde do serviço (1=OK) | <0.99 por 5min |
| **Performance** | `bureau_credito_request_duration_seconds` | Duração das requisições em segundos | P95>500ms por 10min |
| **Tráfego** | `bureau_credito_requests_total` | Total de requisições recebidas | >500/s por 5min |
| **Erros** | `bureau_credito_errors_total` | Total de erros por tipo | >5% de taxa de erro |
| **Saturação** | `bureau_credito_worker_pool_saturation` | Nível de saturação dos workers | >80% por 5min |
| **Compliance** | `bureau_credito_compliance_checks_total` | Total de validações de compliance | >10% de falhas |
| **Negócio** | `bureau_credito_queries_by_market` | Consultas por mercado | N/A |
| **Limites** | `bureau_credito_daily_limit_usage_percent` | Uso do limite diário em percentual | >90% para clientes críticos |

### 4.3 Dashboards Grafana

Os seguintes dashboards Grafana estão disponíveis para monitoramento do módulo:

#### 4.3.1 Dashboard Operacional

![Dashboard Operacional](../../../assets/images/bureau-credito-ops-dashboard.png)

Inclui painéis para:
- Estado geral do serviço
- Latência de requisições (média, P50, P95, P99)
- Taxa de erros
- QPS (Queries por segundo)
- Utilização de recursos (CPU, memória, rede)
- Detalhamento de erros por código e tipo
- Saturation heat map

Acesso: `https://grafana.innovabiz.com/d/bureau-credito-ops`

#### 4.3.2 Dashboard de Negócios

Inclui painéis para:
- Consultas por tipo (Completa, Score, Básica, Restrições)
- Consultas por mercado e finalidade
- Taxa de restrições encontradas
- Distribuição de scores de crédito
- Top clientes por volume de consultas
- Utilização de limites diários
- Tendências de consultas (hora, dia, semana)

Acesso: `https://grafana.innovabiz.com/d/bureau-credito-business`

#### 4.3.3 Dashboard de Compliance

Inclui painéis para:
- Eventos de auditoria por tipo
- Eventos de segurança por severidade
- Falhas de compliance por mercado
- Notificações regulatórias enviadas
- Validações de consentimento
- Taxa de falhas MFA
- Violações de escopo de autorização

Acesso: `https://grafana.innovabiz.com/d/bureau-credito-compliance`

### 4.4 Alertas

Os alertas são gerenciados através do Prometheus Alertmanager e encaminhados para os canais apropriados (email, Slack, PagerDuty) conforme severidade.

#### 4.4.1 Regras de Alertas

```yaml
groups:
- name: bureau-credito-alerts
  rules:
  - alert: BureauCreditoHighErrorRate
    expr: sum(rate(bureau_credito_errors_total[5m])) / sum(rate(bureau_credito_requests_total[5m])) > 0.05
    for: 5m
    labels:
      severity: critical
      service: bureau-credito
    annotations:
      summary: "Alta taxa de erros no Bureau de Crédito"
      description: "Taxa de erro acima de 5% nos últimos 5 minutos ({{ $value | printf \"%.2f\" }})"
      runbook: "https://wiki.innovabiz.com/runbooks/bureau-credito-high-error-rate"

  - alert: BureauCreditoHighLatency
    expr: histogram_quantile(0.95, sum(rate(bureau_credito_request_duration_seconds_bucket[5m])) by (le)) > 0.5
    for: 10m
    labels:
      severity: warning
      service: bureau-credito
    annotations:
      summary: "Alta latência no Bureau de Crédito"
      description: "Latência P95 acima de 500ms nos últimos 10 minutos ({{ $value | printf \"%.2f\" }}s)"
      runbook: "https://wiki.innovabiz.com/runbooks/bureau-credito-high-latency"

  - alert: BureauCreditoComplianceFailure
    expr: sum(increase(bureau_credito_compliance_checks_failed_total[1h])) > 10
    for: 5m
    labels:
      severity: critical
      service: bureau-credito
      domain: compliance
    annotations:
      summary: "Falhas de compliance no Bureau de Crédito"
      description: "Mais de 10 falhas de compliance na última hora"
      runbook: "https://wiki.innovabiz.com/runbooks/bureau-credito-compliance-failures"
      
  - alert: BureauCreditoLimitNearlyExhausted
    expr: bureau_credito_daily_limit_usage_percent{client_tier="premium"} > 90
    for: 5m
    labels:
      severity: warning
      service: bureau-credito
      domain: business
    annotations:
      summary: "Limite diário quase esgotado para cliente premium"
      description: "Cliente {{ $labels.client_id }} está com {{ $value | printf \"%.1f\" }}% do limite diário utilizado"
      runbook: "https://wiki.innovabiz.com/runbooks/bureau-credito-limit-exhaustion"
```

#### 4.4.2 Notificação e Escalonamento

O fluxo de notificações segue a matriz de escalonamento:

| Severidade | Canais Iniciais | Tempo de Escalonamento | Escalonado Para |
|------------|-----------------|------------------------|------------------|
| Critical | Slack #bureau-alerts, Email equipe, PagerDuty | 15 min sem resolução | Gerente de Operações |
| Warning | Slack #bureau-alerts, Email equipe | 30 min sem resolução | PagerDuty |
| Info | Slack #bureau-alerts | N/A | N/A |

### 4.5 Logs

O Bureau de Crédito produz logs estruturados em formato JSON através da biblioteca Zap, configurada para enviar logs para o OpenTelemetry Collector.

#### 4.5.1 Níveis de Log

| Nível | Uso | Exemplo |
|-------|-----|---------|
| DEBUG | Informações detalhadas de desenvolvimento | Valores de parâmetros, estados internos |
| INFO | Eventos normais do sistema | Consulta iniciada, consulta concluída |
| WARN | Situações potencialmente problemáticas | Limite diário próximo do fim, tempo de resposta elevado |
| ERROR | Erros recuperáveis | Falha temporária de integração, timeout |
| FATAL | Erros não recuperáveis | Falha na inicialização, corrupção de configuração crítica |

#### 4.5.2 Estrutura de Logs

```json
{
  "level": "info",
  "timestamp": "2025-08-06T14:30:45.123Z",
  "service": "bureau-credito",
  "version": "1.0.0",
  "environment": "production",
  "market": "brazil",
  "tenant_id": "tenant-123",
  "trace_id": "8a4f301c3ce24c62b34ad2cce58e375f",
  "span_id": "4adf301f3de25c63b67cd2cbf59e785d",
  "consulta_id": "query-123456",
  "entity_id": "entity-abc",
  "message": "Consulta processada com sucesso",
  "elapsed_ms": 235,
  "response_size_bytes": 1458,
  "score": 750,
  "query_type": "ConsultaCompleta",
  "purpose": "FinalidadeConcessaoCredito",
  "has_restrictions": false
}
```

#### 4.5.3 Consulta de Logs

Os logs podem ser consultados através do Kibana:

```
service: "bureau-credito" AND level: "error" AND market: "brazil"
```

Para rastrear uma consulta específica:

```
consulta_id: "query-123456" OR trace_id: "8a4f301c3ce24c62b34ad2cce58e375f"
```

### 4.6 Rastreamento Distribuído

O Bureau de Crédito utiliza OpenTelemetry para instrumentação, produzindo spans para todas as operações significativas.

#### 4.6.1 Spans Principais

| Span | Descrição | Tags/Atributos Importantes |
|------|-----------|----------------------------|
| `bureau.credito.consulta` | Span raiz da consulta | `consulta_id`, `tipo_consulta`, `mercado` |
| `bureau.credito.autenticacao` | Validação de autenticação | `resultado`, `mfa_level` |
| `bureau.credito.autorizacao` | Verificação de autorização | `escopos`, `resultado` |
| `bureau.credito.consentimento` | Validação de consentimento | `consentimento_id`, `valido` |
| `bureau.credito.compliance` | Verificações de compliance | `regras_aplicadas`, `resultado` |
| `bureau.credito.limits` | Verificação de limites diários | `consumido`, `limite` |
| `bureau.credito.core` | Processamento principal | `score`, `restricoes_encontradas` |
| `bureau.credito.notificacao` | Envio de notificações | `tipo_notificacao`, `destinatario` |

#### 4.6.2 Exemplo de Rastreamento

```
bureau.credito.consulta
├── bureau.credito.autenticacao
├── bureau.credito.autorizacao
├── bureau.credito.consentimento
├── bureau.credito.compliance
│   ├── bureau.credito.compliance.angola.bna
│   └── bureau.credito.compliance.global
├── bureau.credito.limits
├── bureau.credito.core
│   ├── bureau.credito.core.score
│   ├── bureau.credito.core.restricoes
│   └── bureau.credito.core.historico
└── bureau.credito.notificacao
```

Acesso ao Jaeger UI: `https://jaeger.innovabiz.com`

## 5. Resolução de Problemas

### 5.1 Troubleshooting Comum

#### 5.1.1 Falhas de Autenticação e Autorização

**Sintomas:**
- Erros HTTP 401 ou 403
- Alto volume de logs de erro relacionados a autenticação
- Alertas `bureau_credito_auth_errors_total`

**Passos de Diagnóstico:**
1. Verificar validade do token JWT nas requisições
2. Confirmar configuração dos escopos necessários no IAM
3. Verificar logs do API Gateway para detalhes de autorização
4. Validar configuração de mercado e tenant

**Resolução:**
```bash
# Verificar configuração de integração IAM
kubectl get configmap bureau-credito-config -n innovabiz -o yaml | grep IAM

# Reiniciar o módulo em caso de problemas de cache de configuração
kubectl rollout restart deployment/bureau-credito -n innovabiz
```

#### 5.1.2 Degradação de Performance

**Sintomas:**
- Latência elevada (P95 > 500ms)
- Alta utilização de CPU/memória
- Alertas `BureauCreditoHighLatency`

**Passos de Diagnóstico:**
1. Verificar métricas de recursos (CPU, memória, rede)
2. Analisar traces para identificar gargalos
3. Verificar volume de consultas por segundo
4. Inspecionar logs para erros externos ou timeouts

**Resolução:**
```bash
# Escalar horizontalmente (curto prazo)
kubectl scale deployment bureau-credito -n innovabiz --replicas=5

# Verificar conexões externas
kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl -v telemetry.example.com

# Verificar estatísticas de GC se alta utilização de memória
kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/debug/pprof/heap > heap.prof
```

#### 5.1.3 Falhas de Compliance

**Sintomas:**
- Erros HTTP 422 com mensagens de compliance
- Alertas `BureauCreditoComplianceFailure`
- Alta taxa de rejeição para consultas específicas

**Passos de Diagnóstico:**
1. Verificar logs de compliance por mercado
2. Confirmar carregamento das regras de compliance
3. Validar configuração específica do mercado
4. Verificar alterações recentes em regulações

**Resolução:**
```bash
# Verificar configuração de compliance carregada
kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/compliance/rules

# Reiniciar carregamento de regras
kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl -X POST http://localhost:8080/internal/compliance/reload
```

#### 5.1.4 Limites Diários Excedidos

**Sintomas:**
- Erros HTTP 429
- Alertas `BureauCreditoLimitNearlyExhausted`
- Reclamações de clientes sobre rejeição de consultas

**Passos de Diagnóstico:**
1. Verificar logs de utilização de limites
2. Identificar padrões anômalos de consumo
3. Verificar configuração de limites por mercado/tenant

**Resolução:**
```bash
# Verificar contadores atuais
kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/limits/counters

# Aumentar limite temporariamente para cliente específico
kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl -X PUT http://localhost:8080/internal/limits/override -d '{"entity_id":"entidade-xyz","new_limit":500}'
```

### 5.2 Matriz de Diagnóstico

| Problema | Sintomas | Verificações Iniciais | Logs/Métricas | Resolução |
|----------|----------|----------------------|---------------|-----------|
| **Lentidão** | Tempo resposta >500ms | CPU, Memória, Traces | `bureau_credito_request_duration_seconds` | Escalar, otimizar caches |
| **Erros 401/403** | Falhas autenticação | Config JWT, IAM | `bureau_credito_auth_errors_total` | Verificar integração IAM |
| **Erros 429** | Limites excedidos | Contadores, Config | `bureau_credito_daily_limit_*` | Ajustar limites, verificar abnormalidades |
| **Erros 422** | Falhas compliance | Regras por mercado | `bureau_credito_compliance_*` | Atualizar regras, verificar dados da requisição |
| **Erros 500** | Falhas internas | Logs, Exceptions | `bureau_credito_errors_*` | Correção de código, reiniciar serviço |

### 5.3 Ferramentas de Diagnóstico

#### 5.3.1 Health Checks

```bash
# Verificar liveness
curl http://<service-ip>:8080/health/live

# Verificar readiness
curl http://<service-ip>:8080/health/ready

# Verificação detalhada (requer autenticação)
curl -H "Authorization: Bearer $TOKEN" http://<service-ip>:8080/health/details
```

#### 5.3.2 Endpoints de Diagnóstico

Os seguintes endpoints estão disponíveis na interface interna (somente dentro do cluster):

```bash
# Estatísticas de memória
curl http://<service-ip>:8080/debug/vars

# Profile de CPU (30s)
curl http://<service-ip>:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Verificação de configuração
curl http://<service-ip>:8080/internal/config/check

# Status dos workers
curl http://<service-ip>:8080/internal/workers/status
```## 6. Procedimentos de Manutenção

### 6.1 Manutenção Programada

A manutenção programada do Bureau de Crédito deve seguir o calendário de janelas de manutenção aprovado, tipicamente:

| Ambiente | Janela Padrão | Frequência | Notificação Prévia |
|----------|---------------|------------|-------------------|
| Desenvolvimento | Qualquer momento | Conforme necessário | Nenhuma |
| Qualidade | Terça, 09:00-12:00 | Semanal | 24 horas |
| Homologação | Quarta, 09:00-12:00 | Quinzenal | 48 horas |
| Produção | Domingo, 01:00-05:00 | Mensal | 7 dias |

#### 6.1.1 Procedimento de Manutenção Padrão

1. **Anúncio**:
   - Criar ticket no Jira (`MAINT-XXX`)
   - Notificar stakeholders via email e Slack
   - Atualizar calendário de manutenção

2. **Preparação**:
   - Verificar estado atual do serviço
   - Certificar-se de que backups recentes estão disponíveis
   - Preparar scripts de rollback
   - Validar atualizações em ambiente de homologação

3. **Execução**:
   - Ativar banner de manutenção através do API Gateway
   - Reduzir tráfego gradualmente (se necessário)
   - Executar procedimentos de manutenção
   - Verificar funcionamento após manutenção
   - Executar testes de sanidade automatizados

4. **Conclusão**:
   - Remover banner de manutenção
   - Notificar conclusão para stakeholders
   - Atualizar ticket de manutenção
   - Documentar quaisquer problemas ou lições aprendidas

### 6.2 Atualização de Regras de Compliance

O procedimento de atualização de regras de compliance por mercado deve ser executado quando houver mudanças regulatórias ou ajustes de negócio:

1. **Preparação das Regras**:
   ```bash
   # Exportar regras atuais
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/compliance/rules > compliance_rules_backup.json
   
   # Preparar arquivo com novas regras
   vim compliance_rules_new.json
   ```

2. **Validação em Homologação**:
   ```bash
   # Aplicar novas regras em homologação
   curl -X POST -H "Content-Type: application/json" -d @compliance_rules_new.json https://bureau-credito-homolog.innovabiz.com/internal/compliance/update
   
   # Executar testes automatizados de compliance
   cd /path/to/tests && go test -v -tags=compliance ./...
   ```

3. **Implantação em Produção**:
   ```bash
   # Aplicar novas regras em produção
   curl -X POST -H "Content-Type: application/json" -d @compliance_rules_new.json https://bureau-credito.innovabiz.com/internal/compliance/update
   
   # Verificar aplicação das regras
   curl https://bureau-credito.innovabiz.com/internal/compliance/rules/status
   ```

### 6.3 Rotação de Segredos

Os segredos do Bureau de Crédito (certificados, tokens, chaves de API) devem ser rotacionados regularmente:

| Segredo | Frequência de Rotação | Procedimento | Impacto |
|---------|----------------------|-------------|---------|
| Certificados TLS | 12 meses | Vault automated | Sem downtime |
| API Keys | 3 meses | Procedimento manual | Possível downtime |
| Tokens de Serviço | 1 mês | Rotação automática | Sem downtime |

Procedimento de rotação manual de API Keys:

1. **Gerar novas credenciais**:
   ```bash
   # Gerar nova API key
   vault write secret/bureau-credito/api-keys/provider-xyz rotation=true
   
   # Recuperar nova API key
   NEW_API_KEY=$(vault read -field=api_key secret/bureau-credito/api-keys/provider-xyz)
   ```

2. **Atualizar serviço**:
   ```bash
   # Atualizar secret no Kubernetes
   kubectl create secret generic bureau-credito-api-keys \
     --from-literal=provider-xyz=$NEW_API_KEY \
     -n innovabiz \
     --dry-run=client -o yaml | kubectl apply -f -
   
   # Reiniciar pods para aplicar nova configuração
   kubectl rollout restart deployment/bureau-credito -n innovabiz
   ```

3. **Verificar funcionamento**:
   ```bash
   # Testar integração com o provedor
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/providers/test?provider=xyz
   ```

4. **Revogar credenciais antigas**:
   ```bash
   # Revogar API key antiga após período de sobreposição
   vault write secret/bureau-credito/api-keys/provider-xyz/revoke confirmed=true
   ```

### 6.4 Gestão de Capacidade

O Bureau de Crédito deve ter sua capacidade monitorada e ajustada regularmente para atender às demandas do negócio:

#### 6.4.1 Monitoramento de Capacidade

| Recurso | Métrica | Limite de Alerta | Limite Crítico |
|---------|---------|------------------|----------------|
| CPU | `bureau_credito_cpu_usage_percent` | 70% por 30min | 85% por 15min |
| Memória | `bureau_credito_memory_usage_percent` | 75% por 30min | 90% por 15min |
| Disco | `bureau_credito_disk_usage_percent` | 75% | 90% |
| Rede | `bureau_credito_network_saturation` | 70% | 85% |
| Requisições | `bureau_credito_requests_per_second` | 80% da capacidade | 90% da capacidade |

#### 6.4.2 Planejamento de Capacidade

A capacidade deve ser revisada mensalmente, considerando:
- Crescimento histórico do volume de consultas
- Previsões de negócios para novos clientes
- Expansão para novos mercados
- Mudanças sazonais (ex: períodos de alto volume)
- Novos tipos de consultas ou funcionalidades

#### 6.4.3 Ajuste de Capacidade

```yaml
# Ajuste do HPA para aumentar capacidade
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: bureau-credito-hpa
  namespace: innovabiz
spec:
  minReplicas: 5  # Aumentado de 3 para 5
  maxReplicas: 15  # Aumentado de 10 para 15
```

```yaml
# Ajuste de recursos dos pods
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bureau-credito
  namespace: innovabiz
spec:
  template:
    spec:
      containers:
      - name: bureau-credito
        resources:
          limits:
            cpu: "2"    # Aumentado de 1 para 2
            memory: "2Gi"  # Aumentado de 1Gi para 2Gi
          requests:
            cpu: "1"    # Aumentado de 500m para 1
            memory: "1Gi"  # Aumentado de 512Mi para 1Gi
```

## 7. Procedimentos de Backup e Recuperação

### 7.1 Estratégia de Backup

O Bureau de Crédito utiliza uma abordagem multicamada para backups:

| Dados | Frequência | Retenção | Método | Responsável |
|-------|------------|----------|--------|------------|
| Configurações | Diário | 90 dias | GitOps (ArgoCD) | DevSecOps |
| Dados Cache | Diário | 7 dias | Redis snapshot | Infraestrutura |
| Métricas | Semanal | 1 ano | Prometheus snapshot | Infraestrutura |
| Logs | Contínuo | 1 ano | Log shipping | Infraestrutura |
| Secrets | Semanal | 1 ano | Vault export | Segurança |

### 7.2 Procedimentos de Backup

#### 7.2.1 Backup de Configurações

As configurações são gerenciadas via GitOps e armazenadas no repositório Git:

```bash
# Verificar status do repositório de configuração
argocd app get bureau-credito

# Exportar configurações atuais para validação
kubectl get cm,secret -l app=bureau-credito -n innovabiz -o yaml > bureau_config_backup.yaml
```

#### 7.2.2 Backup de Cache Redis

```bash
# Iniciar snapshot do Redis
kubectl exec -it redis-master-0 -n cache -- redis-cli SAVE

# Copiar arquivo RDB para armazenamento seguro
kubectl cp cache/redis-master-0:/data/dump.rdb bureau-credito-redis-backup-$(date +%Y%m%d).rdb
```

### 7.3 Procedimentos de Recuperação

#### 7.3.1 Recuperação de Falha de Pod

```bash
# Verificar status dos pods
kubectl get pods -n innovabiz -l app=bureau-credito

# Reiniciar pod específico
kubectl delete pod bureau-credito-5d4f8c7b68-abcd1 -n innovabiz

# Verificar logs do novo pod
kubectl logs -f $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz
```

#### 7.3.2 Recuperação de Configuração

```bash
# Reverter para versão anterior via GitOps
argocd app history bureau-credito
argocd app rollback bureau-credito 15  # Reverter para a versão 15

# Aplicar backup de configuração manualmente (somente emergências)
kubectl apply -f bureau_config_backup.yaml
```

#### 7.3.3 Recuperação de Cache

```bash
# Copiar backup RDB para o pod Redis
kubectl cp bureau-credito-redis-backup-20250805.rdb cache/redis-master-0:/data/dump.rdb.restore

# Restaurar dados dentro do pod
kubectl exec -it redis-master-0 -n cache -- bash
mv /data/dump.rdb.restore /data/dump.rdb
redis-cli SHUTDOWN SAVE
exit
```

### 7.4 Plano de Recuperação de Desastres (DRP)

| Cenário | RTO | RPO | Procedimento |
|---------|-----|-----|-------------|
| Falha de Pod | 5 min | 0 | Recuperação automática via Kubernetes |
| Falha de Nó | 10 min | 0 | Rescheduling automático dos pods |
| Falha de Zona | 30 min | 5 min | Ativação de zona alternativa |
| Falha de Região | 1 hora | 15 min | Failover para região secundária |
| Corrupção de Dados | 2 horas | 24 horas | Restauração a partir de backup |

#### 7.4.1 Failover Regional

```bash
# Verificar status da região secundária
kubectl --context=gcp-europe-west4 get pods -n innovabiz

# Promover região secundária para primária
kubectl --context=gcp-europe-west4 patch configmap global-config -n innovabiz --type merge -p '{"data":{"PRIMARY_REGION":"europe-west4"}}'

# Atualizar DNS para apontar para a região secundária
kubectl --context=gcp-europe-west4 apply -f dns-failover.yaml

# Notificar stakeholders
./scripts/notify-disaster-recovery.sh --event=regional-failover --region=europe-west4
```## 8. Políticas de Segurança Operacional

### 8.1 Princípios de Segurança Operacional

O Bureau de Crédito segue os seguintes princípios de segurança operacional:

1. **Defesa em Profundidade**: Múltiplas camadas de controles de segurança
2. **Princípio do Menor Privilégio**: Acesso mínimo necessário para cada função
3. **Separação de Deveres**: Segregação de responsabilidades para prevenção de fraudes
4. **Segurança por Design**: Controles de segurança incorporados na arquitetura
5. **Zero Trust**: Nenhuma confiança implícita, verificação contínua

### 8.2 Controles de Acesso

#### 8.2.1 Acesso ao Ambiente

| Nível de Acesso | Grupo | Permissões | Método de Autenticação |
|-----------------|-------|------------|------------------------|
| Leitura | bureau-credito-viewers | Visualizar logs, métricas e status | MFA Padrão |
| Operação | bureau-credito-operators | Reiniciar serviço, ajustar configurações | MFA Alto |
| Administração | bureau-credito-admins | Acesso completo, incluindo secrets | MFA Alto + Aprovação |

#### 8.2.2 Rotação de Credenciais

- Credenciais de serviço: Rotação automática a cada 30 dias
- Credenciais de usuário: Expiração em 90 dias
- Credenciais emergenciais: Expiração em 24 horas

#### 8.2.3 Gestão de Sessões

- Timeout de sessão: 30 minutos de inatividade
- Duração máxima de sessão: 8 horas
- Bloqueio após 5 tentativas malsucedidas

### 8.3 Segurança de Dados

#### 8.3.1 Classificação de Dados

| Classificação | Exemplos | Controles |
|---------------|----------|-----------|
| Público | Documentação pública | Sem restrições |
| Interno | Configurações não-sensíveis | Autenticação básica |
| Confidencial | Dados de consulta, respostas | Criptografia, acesso controlado |
| Restrito | Documentos pessoais, dados financeiros | Criptografia forte, MFA alto, mascaramento |

#### 8.3.2 Criptografia

- Dados em trânsito: TLS 1.3
- Dados em repouso: AES-256
- Chaves sensíveis: Gerenciadas via Vault com HSM

#### 8.3.3 Mascaramento de Dados

```go
// Exemplo de implementação do mascaramento de dados sensíveis
func maskSensitiveData(data *ResultadoConsulta, market string) {
    switch market {
    case "brazil":
        // CPF: mostra apenas os 3 últimos dígitos
        if data.DocumentoCliente != "" {
            data.DocumentoCliente = "***.***.***-" + data.DocumentoCliente[len(data.DocumentoCliente)-2:]
        }
        
        // Dados financeiros: valores exatos apenas com escopo especial
        if !hasScope(ctx, "bureau_credito:dados_financeiros:completo") {
            for i := range data.RegistrosCredito {
                data.RegistrosCredito[i].Valor = roundToRange(data.RegistrosCredito[i].Valor)
            }
        }
    case "eu":
        // GDPR: remoção de dados pessoais sem consentimento explícito
        if !hasConsent(ctx, "personal_data_processing") {
            data.EnderecoCompleto = ""
            data.Telefones = nil
        }
    }
    // Demais regras por mercado...
}
```

### 8.4 Segurança de Comunicações

#### 8.4.1 Política de Rede

- Segmentação de rede via Network Policies
- Tráfego entre pods criptografado via Service Mesh
- Exposição externa somente via API Gateway
- Bloqueio de tráfego egress não autorizado

```yaml
# Network Policy para Bureau de Crédito
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: bureau-credito-network-policy
  namespace: innovabiz
spec:
  podSelector:
    matchLabels:
      app: bureau-credito
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-proxy
    - podSelector:
        matchLabels:
          app: api-gateway
    - podSelector:
        matchLabels:
          app: otel-collector
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
  - to:
    - podSelector:
        matchLabels:
          app: iam-service
    ports:
    - protocol: TCP
      port: 8080
  - to:
    - podSelector:
        matchLabels:
          app: otel-collector
    ports:
    - protocol: TCP
      port: 4317
```

#### 8.4.2 API Security

- Validação de entrada via JSON Schema
- Proteção contra ataques de injeção
- Rate limiting por cliente
- Proteção OWASP Top 10

### 8.5 Gestão de Vulnerabilidades

#### 8.5.1 Ciclo de Vida de Patches

| Severidade | SLA para Aplicação | Janela de Manutenção |
|------------|-------------------|----------------------|
| Crítica | 24 horas | Imediata |
| Alta | 7 dias | Próxima janela |
| Média | 30 dias | Janela mensal |
| Baixa | 90 dias | Janela trimestral |

#### 8.5.2 Processo de Gestão

```bash
# Verificar vulnerabilidades da imagem
trivy image innovabiz/bureau-credito:latest

# Verificar vulnerabilidades do deployment
kubectl-trivy deployment bureau-credito -n innovabiz

# Gerar relatório de conformidade
trivy image --format json --output bureau-credito-vulnerabilities.json innovabiz/bureau-credito:latest
```

## 9. Conformidade por Mercado

### 9.1 Angola (BNA)

#### 9.1.1 Requisitos Regulatórios

- **Base Legal**: Aviso nº 05/2021 do Banco Nacional de Angola (BNA)
- **Escopo**: Instituições financeiras bancárias e não bancárias
- **Principais Requisitos**:
  - Consentimento explícito para consultas
  - Notificação obrigatória para todas as consultas
  - MFA de nível alto para consultas completas
  - Retenção de registros por 5 anos

#### 9.1.2 Configuração Específica

```yaml
# ConfigMap específico para Angola
apiVersion: v1
kind: ConfigMap
metadata:
  name: bureau-credito-angola-config
  namespace: innovabiz
data:
  CONSENT_REQUIRED: "true"
  NOTIFICATION_REQUIRED: "true"
  MIN_MFA_LEVEL_CONSULTA_COMPLETA: "high"
  MIN_MFA_LEVEL_CONSULTA_SCORE: "standard"
  RETENTION_PERIOD_DAYS: "1825" # 5 anos
  DAILY_QUERY_LIMIT: "50"
  MASKING_RULES: |
    {
      "DocumentoCliente": "mask-partial",
      "ValorCredito": "no-mask",
      "Endereco": "mask-partial"
    }
```

#### 9.1.3 Procedimentos de Auditoria

Procedimentos a serem executados trimestralmente:

1. **Verificação de Consentimentos**:
   ```bash
   # Exportar logs de consentimento para análise
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl -o /tmp/consent-audit.json http://localhost:8080/internal/audit/consent?market=angola&startDate=2025-05-01
   
   # Analisar taxas de conformidade
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/compliance/stats?market=angola | jq '.consentStats'
   ```

2. **Verificação de Notificações**:
   ```bash
   # Verificar registros de notificações
   kubectl exec -it $(kubectl get pods -n innovabiz -l app=bureau-credito -o jsonpath='{.items[0].metadata.name}') -n innovabiz -- curl http://localhost:8080/internal/notifications/stats?market=angola
   ```

### 9.2 Brasil (BACEN/LGPD)

#### 9.2.1 Requisitos Regulatórios

- **Base Legal**: Lei Geral de Proteção de Dados (LGPD), Resolução nº 4,737 do Banco Central do Brasil
- **Escopo**: Instituições financeiras, birôs de crédito
- **Principais Requisitos**:
  - Finalidade específica para cada consulta
  - Notificação obrigatória para restrições de crédito
  - Consentimento para compartilhamento de dados pessoais
  - Direito de acesso, correção e exclusão de dados

#### 9.2.2 Configuração Específica

```yaml
# ConfigMap específico para Brasil
apiVersion: v1
kind: ConfigMap
metadata:
  name: bureau-credito-brasil-config
  namespace: innovabiz
data:
  PURPOSE_REQUIRED: "true"
  NOTIFICATION_FOR_RESTRICTIONS: "true"
  CONSENT_REQUIRED: "true"
  MIN_MFA_LEVEL_CONSULTA_COMPLETA: "standard"
  RETENTION_PERIOD_DAYS: "730" # 2 anos
  DATA_SUBJECT_RIGHTS_ENABLED: "true"
  DAILY_QUERY_LIMIT: "200"
  LGPD_DATA_CATEGORIES: |
    {
      "dados_pessoais": ["nome", "cpf", "endereco", "telefone"],
      "dados_financeiros": ["score", "restricoes", "historico_credito"],
      "dados_sensiveis": []
    }
```

### 9.3 União Europeia (GDPR/PSD2)

#### 9.3.1 Requisitos Regulatórios

- **Base Legal**: General Data Protection Regulation (GDPR), Payment Services Directive 2 (PSD2)
- **Escopo**: Entidades que processam dados de cidadãos da UE
- **Principais Requisitos**:
  - Minimização de dados para finalidade específica
  - Consentimento explícito e específico
  - Direito ao esquecimento
  - Notificação de violação de dados em 72h

#### 9.3.2 Configuração Específica

```yaml
# ConfigMap específico para EU
apiVersion: v1
kind: ConfigMap
metadata:
  name: bureau-credito-eu-config
  namespace: innovabiz
data:
  DATA_MINIMIZATION: "true"
  EXPLICIT_CONSENT_REQUIRED: "true"
  RIGHT_TO_BE_FORGOTTEN_ENABLED: "true"
  DATA_BREACH_NOTIFICATION_ENABLED: "true"
  MIN_MFA_LEVEL_CONSULTA_COMPLETA: "high"
  RETENTION_PERIOD_DAYS: "365" # 1 ano
  CONSENT_EXPIRY_DAYS: "90" # Expiração em 90 dias
  DATA_PORTABILITY_ENABLED: "true"
  DAILY_QUERY_LIMIT: "100"
```

### 9.4 EUA (FCRA/GLBA)

#### 9.4.1 Requisitos Regulatórios

- **Base Legal**: Fair Credit Reporting Act (FCRA), Gramm-Leach-Bliley Act (GLBA)
- **Escopo**: Birôs de crédito, instituições financeiras
- **Principais Requisitos**:
  - Finalidade permissível para consultas
  - Notificação de decisão adversa
  - Direito de contestação
  - Segurança de dados financeiros

#### 9.4.2 Configuração Específica

```yaml
# ConfigMap específico para USA
apiVersion: v1
kind: ConfigMap
metadata:
  name: bureau-credito-usa-config
  namespace: innovabiz
data:
  PERMISSIBLE_PURPOSE_REQUIRED: "true"
  ADVERSE_ACTION_NOTIFICATION: "true"
  DISPUTE_RESOLUTION_ENABLED: "true"
  SAFEGUARDS_RULE_COMPLIANCE: "true"
  MIN_MFA_LEVEL_CONSULTA_COMPLETA: "standard"
  RETENTION_PERIOD_DAYS: "2555" # 7 anos
  IDENTITY_THEFT_PROTECTION: "true"
  DAILY_QUERY_LIMIT: "150"
```

## 10. Gestão de Configuração

### 10.1 Estratégia de Configuração

O Bureau de Crédito utiliza uma abordagem em camadas para configuração:

1. **Configuração Base**: Aplicada a todos os ambientes e mercados
2. **Configuração por Ambiente**: Sobrescreve configurações base por ambiente (dev, qa, homolog, prod)
3. **Configuração por Mercado**: Sobrescreve configurações por mercado (angola, brasil, eu, usa, global)
4. **Configuração por Tenant**: Ajustes específicos por tenant quando necessário

### 10.2 Fontes de Configuração

| Prioridade | Fonte | Tipo | Uso |
|------------|-------|------|-----|
| 1 (menor) | Padrões hardcoded | Valores em código | Defaults de última instância |
| 2 | ConfigMap base | Valores globais | Configurações comuns |
| 3 | ConfigMap ambiente | Valores por ambiente | Endpoints de serviços |
| 4 | ConfigMap mercado | Valores por mercado | Regras específicas |
| 5 | Secret | Valores sensíveis | Credenciais, chaves |
| 6 (maior) | Variáveis de ambiente | Override | Configurações de emergência |

### 10.3 Chaves de Configuração

| Chave | Descrição | Valores Padrão | Ambientes Aplicáveis |
|-------|-----------|----------------|----------------------|
| `ENVIRONMENT` | Ambiente de execução | `production` | Todos |
| `MARKET` | Mercado padrão | `global` | Todos |
| `TENANT_TYPE` | Tipo de tenant | `default` | Todos |
| `SERVICE_VERSION` | Versão do serviço | `1.0.0` | Todos |
| `LOG_LEVEL` | Nível de log | `info` | Todos |
| `DAILY_QUERY_LIMIT_*` | Limite diário por mercado | Varia | Todos |
| `NOTIFICATION_REQUIRED_*` | Requisitos de notificação | Varia | Todos |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Endpoint do coletor OpenTelemetry | `http://otel-collector:4317` | Todos |
| `REDIS_HOST` | Host do Redis para cache | `redis-master.cache` | Todos |
| `REDIS_PORT` | Porta do Redis | `6379` | Todos |
| `IAM_SERVICE_URL` | URL do serviço IAM | Varia | Todos |
| `COMPLIANCE_RULES_PATH` | Caminho do arquivo de regras | `/app/config/compliance-rules.json` | Todos |
| `MFA_REQUIRED_LEVELS` | Níveis de MFA por operação | JSON com configurações | Todos |

### 10.4 Gestão de Secrets

O Bureau de Crédito utiliza o HashiCorp Vault integrado ao Kubernetes para gerenciar secrets:

```bash
# Verificar status dos secrets
kubectl get secrets -n innovabiz -l app=bureau-credito

# Rotacionar secrets
vault write -f secret/bureau-credito/rotate

# Sincronizar secrets com Kubernetes
vault-k8s-sync bureau-credito
```

## 11. Referências

### 11.1 Documentação Interna

- [ADR: Bureau de Crédito](../adr/bureau-credito-adr.md)
- [Especificação Técnica](../technical/bureau-credito-technical-spec.md)
- [Guia de Integração](../integration/bureau-credito-integration-guide.md)
- [Plano de Testes](../testing/bureau-credito-test-plan.md)
- [Runbooks](../runbooks/)

### 11.2 Documentação de Ferramentas

- [Kubernetes](https://kubernetes.io/docs/)
- [OpenTelemetry](https://opentelemetry.io/docs/)
- [Prometheus](https://prometheus.io/docs/)
- [Grafana](https://grafana.com/docs/)
- [KrakenD API Gateway](https://www.krakend.io/docs/)
- [HashiCorp Vault](https://www.vaultproject.io/docs)

### 11.3 Regulamentações

- [BNA - Aviso nº 05/2021](https://www.bna.ao/)
- [BACEN - Resolução nº 4,737](https://www.bcb.gov.br/)
- [LGPD - Lei Geral de Proteção de Dados](https://www.lgpdbrasil.com.br/)
- [GDPR](https://gdpr.eu/)
- [PSD2](https://ec.europa.eu/info/law/payment-services-psd-2-directive-eu-2015-2366_en)
- [FCRA](https://www.ftc.gov/enforcement/statutes/fair-credit-reporting-act)
- [GLBA](https://www.ftc.gov/business-guidance/privacy-security/gramm-leach-bliley-act)

---

**Histórico de Revisões**

| Versão | Data | Autor | Descrição |
|--------|------|-------|-----------|
| 0.1 | 2025-07-15 | Equipe de Operações | Versão inicial |
| 0.2 | 2025-07-28 | Equipe de Segurança | Adição de políticas de segurança |
| 0.3 | 2025-08-01 | Equipe de Compliance | Adição de requisitos regulatórios |
| 1.0 | 2025-08-06 | Equipe de Operações | Versão final aprovada |

**Classificação: Privado - Uso Interno**