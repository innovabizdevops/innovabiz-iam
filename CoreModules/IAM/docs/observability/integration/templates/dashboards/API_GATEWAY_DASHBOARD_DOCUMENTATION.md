# Documentação Técnica: Dashboard API Gateway INNOVABIZ

![Status](https://img.shields.io/badge/Status-Implementado-brightgreen)
![Versão](https://img.shields.io/badge/Versão-1.0.0-blue)
![Plataforma](https://img.shields.io/badge/Plataforma-INNOVABIZ-purple)
![Compliance](https://img.shields.io/badge/Compliance-ISO%2027001%20|%20PCI%20DSS%20|%20LGPD-orange)

## 1. Visão Geral

### 1.1 Objetivos do Dashboard

O dashboard API Gateway INNOVABIZ foi desenvolvido para fornecer monitoramento abrangente do KrakenD API Gateway, componente crítico da arquitetura de microserviços da plataforma. Este gateway atua como ponto central de entrada para todas as APIs, fornecendo funcionalidades essenciais como roteamento, autenticação, rate limiting, circuit breakers e agregação de dados.

Este dashboard permite:
- Monitorar a saúde e disponibilidade do API Gateway
- Visualizar métricas de tráfego, latência e taxa de erros
- Acompanhar performance de endpoints individuais e backends
- Identificar gargalos de performance e problemas de conectividade
- Observar comportamento de circuit breakers e mecanismos de proteção
- Manter visibilidade completa sobre o tráfego de APIs através dos contextos da plataforma

### 1.2 Arquitetura Multi-Contexto

O dashboard segue rigorosamente a arquitetura multi-contexto da plataforma INNOVABIZ, permitindo filtrar todas as métricas por:

- **Tenant (tenant_id)**: Isolamento por tenant para ambientes multi-tenant
- **Região (region_id)**: Divisão geográfica para conformidade e baixa latência
- **Ambiente (environment)**: Segmentação por ambientes (prod, stage, dev, sandbox)
- **Instância (instance)**: Seleção de instâncias específicas do gateway

Esta arquitetura garante que diferentes stakeholders possam visualizar métricas relevantes para seus respectivos contextos operacionais, mantendo conformidade com requisitos de governança, segregação de dados e isolamento multi-tenant.

## 2. Estrutura do Dashboard

### 2.1 Seções Principais

O dashboard está organizado em seções lógicas para facilitar o monitoramento operacional e troubleshooting:

1. **Status e Disponibilidade**
   - Status da instância (online/offline)
   - Uptime do serviço
   - Taxa de requisições total
   - Taxa de sucesso (SLI)

2. **Performance e Latência**
   - Taxa de requisições por status code
   - Latência de requisições (P50, P95, P99)
   - Taxa de requisições por endpoint
   - Latência de backends (P95)

3. **Recursos e Sistema**
   - Utilização de CPU e memória
   - Estado dos circuit breakers
   - Clientes conectados
   - Número de goroutines
   - Uso de memória do processo

4. **Rede e Conectividade**
   - Throughput de rede (RX/TX)
   - Taxa de requisições por método HTTP
   - Métricas de backends

### 2.2 Anotações Automáticas

O dashboard inclui anotações automáticas para eventos críticos:

- **Erros HTTP 5xx**: Detecta aumentos em erros de servidor
- **Circuit Breaker Ativado**: Identifica quando circuit breakers mudam de estado

### 2.3 Variáveis e Filtros

O dashboard implementa variáveis cascateadas para filtragem multi-contexto:

| Variável | Descrição | Dependências |
|----------|-----------|--------------|
| tenant_id | ID do tenant | Nenhuma |
| region_id | ID da região | tenant_id |
| environment | Ambiente (prod, stage, dev, etc.) | tenant_id, region_id |
| instance | Instância específica do KrakenD | tenant_id, region_id, environment |

As variáveis são configuradas para permitir seleção múltipla e incluir a opção "All" (todos), facilitando a navegação de contextos gerais para específicos.

## 3. Requisitos e Implementação

### 3.1 Métricas Prometheus Necessárias

O dashboard requer as seguintes métricas exportadas pelo KrakenD:

**Métricas básicas:**
- `up{job="krakend"}`: Status da instância (0/1)
- `process_start_time_seconds`: Tempo de início do processo
- `krakend_router_connected_clients`: Clientes conectados

**Métricas de requisições:**
- `krakend_http_request_duration_seconds_count`: Contador de requisições
- `krakend_http_request_duration_seconds_sum`: Soma da duração das requisições
- `krakend_http_request_duration_seconds_bucket`: Histograma de latência

**Métricas de backend:**
- `krakend_backend_request_duration_seconds_count`: Requisições para backends
- `krakend_backend_request_duration_seconds_bucket`: Latência de backends

**Métricas de circuit breaker:**
- `krakend_circuit_breaker_state`: Estado dos circuit breakers
- `krakend_circuit_breaker_state_changes_total`: Mudanças de estado

**Métricas de sistema:**
- `node_cpu_seconds_total`: Utilização de CPU
- `node_memory_*`: Métricas de memória
- `node_network_*`: Métricas de rede
- `process_resident_memory_bytes`: Memória do processo
- `go_goroutines`: Número de goroutines

### 3.2 Requisitos de Labels

Para compatibilidade com a arquitetura multi-contexto INNOVABIZ, todas as métricas devem incluir os seguintes labels:

```yaml
- tenant_id: "<identificador_do_tenant>"
- region_id: "<identificador_da_região>"
- environment: "<ambiente>"
- instance: "<host>:<porta>"
```

Labels adicionais específicos do KrakenD:
```yaml
- status_code: "<código_http>"
- endpoint: "<endpoint_path>"
- backend: "<backend_name>"
- method: "<http_method>"
- state: "<circuit_breaker_state>"
```

### 3.3 Configuração do KrakenD

#### 3.3.1 Configuração de Métricas

O KrakenD deve ser configurado para exportar métricas Prometheus:

```json
{
  "version": 3,
  "extra_config": {
    "telemetry/metrics": {
      "collection_time": "60s",
      "proxy_disabled": false,
      "router_disabled": false,
      "backend_disabled": false,
      "endpoint_disabled": false,
      "listen_address": ":8090"
    },
    "telemetry/opencensus": {
      "sample_rate": 100,
      "reporting_period": 1,
      "enabled_layers": {
        "backend": true,
        "router": true,
        "pipe": true
      },
      "exporters": {
        "prometheus": {
          "port": 8090,
          "namespace": "krakend",
          "tag_host": false,
          "tag_path": true,
          "tag_method": true,
          "tag_statuscode": true
        }
      }
    }
  }
}
```

#### 3.3.2 Configuração do Prometheus

Adicione o seguinte scrape config ao Prometheus:

```yaml
scrape_configs:
  - job_name: 'krakend'
    scrape_interval: 15s
    metrics_path: '/metrics'
    static_configs:
      - targets: ['krakend:8090']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
      - source_labels: [__meta_kubernetes_pod_label_tenant_id]
        target_label: tenant_id
      - source_labels: [__meta_kubernetes_pod_label_region_id]
        target_label: region_id
      - source_labels: [__meta_kubernetes_pod_label_environment]
        target_label: environment
```

### 3.4 Configuração do Grafana

Para importar o dashboard:

1. Acesse Grafana > Dashboards > Import
2. Carregue o arquivo JSON do dashboard
3. Selecione a fonte de dados Prometheus
4. Configure permissões conforme políticas de RBAC da plataforma INNOVABIZ
5. Salve o dashboard

## 4. Casos de Uso Operacional

### 4.1 Monitoramento em Tempo Real

**Cenário**: Acompanhamento contínuo do tráfego de APIs e saúde do gateway

**Painéis relevantes**:
- Status da instância
- Taxa de requisições total
- Taxa de requisições por status code
- Latência de requisições

**Procedimento**:
1. Verifique o status online/offline das instâncias
2. Observe a taxa total de requisições para detectar padrões anormais
3. Monitore a distribuição de status codes para identificar problemas
4. Configure o intervalo de atualização automática para 30 segundos

### 4.2 Troubleshooting de Performance

**Cenário**: Investigação de lentidão reportada em APIs

**Painéis relevantes**:
- Latência de requisições (P95, P99)
- Latência de backends
- Taxa de requisições por endpoint
- Utilização de CPU e memória

**Procedimento**:
1. Analise os percentis de latência para determinar se o problema é generalizado
2. Identifique endpoints específicos com alta latência
3. Correlacione com latência de backends para identificar gargalos
4. Verifique se há saturação de recursos (CPU/memória)
5. Analise o estado dos circuit breakers para backends problemáticos

### 4.3 Análise de Segurança

**Cenário**: Investigação de possíveis ataques ou comportamento anômalo

**Painéis relevantes**:
- Taxa de requisições por status code (401, 403, 429)
- Taxa de requisições por endpoint
- Clientes conectados
- Taxa de requisições por método HTTP

**Procedimento**:
1. Observe picos anormais em códigos 401 (não autorizado) e 403 (proibido)
2. Identifique endpoints com tráfego incomum
3. Verifique se há aumento súbito no número de clientes conectados
4. Analise a distribuição de métodos HTTP para detectar padrões suspeitos
5. Correlacione com logs de auditoria para investigação detalhada

### 4.4 Planejamento de Capacidade

**Cenário**: Avaliar necessidade de expansão de recursos para o API Gateway

**Painéis relevantes**:
- Taxa de requisições total (tendências)
- Utilização de CPU e memória
- Throughput de rede
- Número de goroutines

**Procedimento**:
1. Analise tendências de crescimento na taxa de requisições
2. Verifique se CPU está consistentemente acima de 70% ou memória acima de 80%
3. Observe tendências no throughput de rede
4. Monitore crescimento no número de goroutines
5. Utilize período de visualização de 7-30 dias para identificar tendências

## 5. Governança e Compliance

### 5.1 Requisitos de Segurança

O dashboard foi projetado considerando requisitos de segurança em conformidade com:

- **ISO 27001**: Controles de acesso e monitoramento de ativos de informação
- **PCI DSS**: Requisitos 10.1-10.3 para rastreamento de atividades e 6.5 para desenvolvimento seguro
- **GDPR/LGPD**: Segregação de dados por tenant e região para conformidade com legislações de privacidade

### 5.2 Controle de Acesso

Recomenda-se a seguinte matriz de controle de acesso ao dashboard:

| Perfil | Permissão | Escopo |
|--------|-----------|--------|
| Operador NOC | Visualização | Todos os tenants/regiões |
| SRE/DevOps | Visualização | Todos os tenants/regiões |
| Admin Tenant | Visualização | Tenant específico |
| Analista de Segurança | Visualização | Todos os tenants/regiões |
| Dev/QA | Visualização | Apenas ambientes não-produtivos |
| Arquiteto de APIs | Visualização | Todos os tenants/regiões |

### 5.3 Auditoria e Rastreabilidade

Todas as interações com o dashboard devem ser registradas no sistema de auditoria centralizado, incluindo:

- Quem acessou o dashboard
- Quais filtros foram aplicados
- Quando o acesso ocorreu
- Ações realizadas (exportação de dados, configurações alteradas)

## 6. Alertas Recomendados

Baseado nas métricas visualizadas neste dashboard, recomenda-se configurar os seguintes alertas no Prometheus AlertManager:

### 6.1 Alertas de Disponibilidade

```yaml
- alert: APIGatewayDown
  expr: up{job="krakend"} == 0
  for: 1m
  labels:
    severity: critical
    category: availability
  annotations:
    summary: "API Gateway Down"
    description: "A instância {{ $labels.instance }} está offline por pelo menos 1 minuto"
```

### 6.2 Alertas de Performance

```yaml
- alert: APIGatewayHighLatency
  expr: histogram_quantile(0.95, sum(rate(krakend_http_request_duration_seconds_bucket[5m])) by (le)) > 1
  for: 5m
  labels:
    severity: warning
    category: performance
  annotations:
    summary: "Alta Latência P95"
    description: "Latência P95 acima de 1 segundo por 5 minutos"
```

### 6.3 Alertas de Segurança

```yaml
- alert: APIGatewayHighErrorRate
  expr: sum(rate(krakend_http_request_duration_seconds_count{status_code=~"5.."}[5m])) / sum(rate(krakend_http_request_duration_seconds_count[5m])) > 0.05
  for: 5m
  labels:
    severity: warning
    category: security
  annotations:
    summary: "Alta Taxa de Erros 5xx"
    description: "Taxa de erros 5xx acima de 5% nos últimos 5 minutos"
```

## 7. Integração com Outros Dashboards

Este dashboard se integra com outros dashboards da plataforma INNOVABIZ através das seguintes conexões:

### 7.1 Dashboards Relacionados

- **Dashboard de Infraestrutura**: Para correlacionar métricas de infraestrutura com performance do gateway
- **Dashboard de Backend Services**: Para correlacionar latência de APIs com performance de backends
- **Dashboard Alerting**: Para visualização consolidada de alertas relacionados ao API Gateway
- **Dashboard Multi-Contexto**: Para visão holística de todos os componentes por tenant/região

### 7.2 Navegação Entre Dashboards

Links diretos são fornecidos no dashboard para navegação contextualizada entre sistemas relacionados:

- Link para dashboard de infraestrutura com contexto da instância atual
- Link para dashboard de alertas filtrado para alertas do API Gateway
- Link para logs de sistema relacionados às instâncias do gateway

## 8. Melhorias Futuras

### 8.1 Próximas Iterações

- Adicionar métricas específicas de cache (hit/miss ratio)
- Implementar métricas de JWT validation e rate limiting
- Integrar métricas de qualidade de serviço por tenant
- Expandir visualizações de circuit breakers por backend
- Adicionar métricas de transformação de dados e agregação

### 8.2 Integrações Planejadas

- Integração com sistema de tracing distribuído para correlação de requisições
- Implementação de deep-links para logs específicos de requisições
- Correlação automática com incidentes
- Análise preditiva de tendências de tráfego e performance
- Integração com sistema de documentação de APIs

## 9. Referências e Recursos Adicionais

- [Documentação Oficial KrakenD](https://www.krakend.io/docs/)
- [Guia de Observabilidade INNOVABIZ](https://wiki.innovabiz.com/observability-guide)
- [Especificação do KrakenD Telemetry](https://www.krakend.io/docs/telemetry/)
- [RFC INNOVABIZ: Padrões de Monitoramento Multi-Contexto](https://wiki.innovabiz.com/rfc/monitoring-standards)
- [Requisitos de Governança INNOVABIZ](https://wiki.innovabiz.com/governance)
- [API Gateway Best Practices](https://wiki.innovabiz.com/api-gateway-best-practices)

---

**Autor**: Equipe de Plataforma INNOVABIZ  
**Criado**: Fevereiro 2025  
**Última Atualização**: Fevereiro 2025  
**Revisão Programada**: Agosto 2025  
**Classificação**: Interno - Confidencial