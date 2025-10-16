# Template para Dashboards Grafana - INNOVABIZ

## Visão Geral

Este documento fornece o template padrão INNOVABIZ para criação de dashboards Grafana, garantindo consistência visual, funcional e de experiência do usuário em toda a plataforma. Os dashboards seguem os princípios multi-dimensionais da plataforma INNOVABIZ, permitindo visualização e filtragem por tenant, região, módulo e outros contextos relevantes.

## Estrutura de Dashboards

A plataforma INNOVABIZ organiza dashboards Grafana em uma hierarquia padronizada:

1. **Dashboards de Visão Executiva** - Resumo de alto nível para gestores e executivos
2. **Dashboards Operacionais** - Visão detalhada para operações diárias
3. **Dashboards de Troubleshooting** - Detalhes aprofundados para diagnóstico de problemas
4. **Dashboards de Capacidade** - Análise de tendências e planejamento de capacidade

Cada módulo deve implementar estas categorias de dashboards, adaptando métricas e visualizações específicas às suas necessidades.

## Template JSON Base

Abaixo está um template JSON básico para dashboards Grafana compatível com Grafana 9.0+. Este template inclui variáveis de contexto multi-dimensional, layout padronizado e temas consistentes com as diretrizes INNOVABIZ:

```json
{
  "__inputs": [],
  "__elements": {},
  "__requires": [
    {
      "type": "grafana",
      "id": "grafana",
      "name": "Grafana",
      "version": "9.5.0"
    },
    {
      "type": "panel",
      "id": "timeseries",
      "name": "Time series",
      "version": ""
    },
    {
      "type": "panel",
      "id": "stat",
      "name": "Stat",
      "version": ""
    },
    {
      "type": "panel",
      "id": "gauge",
      "name": "Gauge",
      "version": ""
    },
    {
      "type": "panel",
      "id": "table",
      "name": "Table",
      "version": ""
    },
    {
      "type": "datasource",
      "id": "prometheus",
      "name": "Prometheus",
      "version": "1.0.0"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": {
          "type": "grafana",
          "uid": "-- Grafana --"
        },
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "target": {
          "limit": 100,
          "matchAny": false,
          "tags": [],
          "type": "dashboard"
        },
        "type": "dashboard"
      },
      {
        "datasource": {
          "type": "prometheus",
          "uid": "${DS_PROMETHEUS}"
        },
        "enable": true,
        "expr": "changes(version_info{service_name=\"${service}\", innovabiz_tenant_id=\"${tenant}\", innovabiz_region_id=\"${region}\"}[1m]) > 0",
        "iconColor": "#5794F2",
        "name": "Deployments",
        "showIn": 0,
        "tags": ["deployment", "version"]
      }
    ]
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 1,
  "id": null,
  "links": [
    {
      "asDropdown": false,
      "icon": "dashboard",
      "includeVars": true,
      "keepTime": true,
      "tags": ["innovabiz"],
      "targetBlank": false,
      "title": "Painel Central INNOVABIZ",
      "tooltip": "",
      "type": "link",
      "url": "/d/innovabiz-home"
    },
    {
      "asDropdown": true,
      "icon": "external link",
      "includeVars": true,
      "keepTime": true,
      "tags": ["${service}"],
      "targetBlank": false,
      "title": "Outros dashboards ${service}",
      "tooltip": "",
      "type": "dashboards"
    }
  ],
  "liveNow": false,
  "panels": [
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "id": 1,
      "panels": [],
      "title": "Visão Geral",
      "type": "row"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "palette-classic"
          },
          "custom": {
            "axisCenteredZero": false,
            "axisColorMode": "text",
            "axisLabel": "",
            "axisPlacement": "auto",
            "barAlignment": 0,
            "drawStyle": "line",
            "fillOpacity": 10,
            "gradientMode": "none",
            "hideFrom": {
              "legend": false,
              "tooltip": false,
              "viz": false
            },
            "lineInterpolation": "linear",
            "lineWidth": 1,
            "pointSize": 5,
            "scaleDistribution": {
              "type": "linear"
            },
            "showPoints": "never",
            "spanNulls": true,
            "stacking": {
              "group": "A",
              "mode": "none"
            },
            "thresholdsStyle": {
              "mode": "off"
            }
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "red",
                "value": 80
              }
            ]
          },
          "unit": "reqps"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 12,
        "x": 0,
        "y": 1
      },
      "id": 2,
      "options": {
        "legend": {
          "calcs": ["mean", "max", "lastNotNull"],
          "displayMode": "table",
          "placement": "bottom",
          "showLegend": true
        },
        "tooltip": {
          "mode": "multi",
          "sort": "none"
        }
      },
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "editorMode": "code",
          "expr": "sum(rate(http_server_requests_seconds_count{service_name=\"${service}\", innovabiz_tenant_id=\"${tenant}\", innovabiz_region_id=\"${region}\"}[$__rate_interval])) by (status_code)",
          "instant": false,
          "legendFormat": "{{status_code}}",
          "range": true,
          "refId": "A"
        }
      ],
      "title": "Taxa de Requisições por Status",
      "type": "timeseries"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "thresholds"
          },
          "mappings": [],
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "yellow",
                "value": 500
              },
              {
                "color": "orange",
                "value": 1000
              },
              {
                "color": "red",
                "value": 2000
              }
            ]
          },
          "unit": "ms"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 12,
        "y": 1
      },
      "id": 3,
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "justifyMode": "auto",
        "orientation": "auto",
        "reduceOptions": {
          "calcs": ["lastNotNull"],
          "fields": "",
          "values": false
        },
        "textMode": "auto"
      },
      "pluginVersion": "9.5.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "editorMode": "code",
          "expr": "histogram_quantile(0.95, sum(rate(http_server_requests_seconds_bucket{service_name=\"${service}\", innovabiz_tenant_id=\"${tenant}\", innovabiz_region_id=\"${region}\"}[$__rate_interval])) by (le))*1000",
          "instant": false,
          "legendFormat": "p95",
          "range": true,
          "refId": "A"
        }
      ],
      "title": "Latência p95",
      "type": "stat"
    },
    {
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "thresholds"
          },
          "mappings": [],
          "max": 100,
          "min": 0,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              {
                "color": "green",
                "value": null
              },
              {
                "color": "yellow",
                "value": 70
              },
              {
                "color": "orange",
                "value": 85
              },
              {
                "color": "red",
                "value": 95
              }
            ]
          },
          "unit": "percent"
        },
        "overrides": []
      },
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 18,
        "y": 1
      },
      "id": 4,
      "options": {
        "orientation": "auto",
        "reduceOptions": {
          "calcs": ["lastNotNull"],
          "fields": "",
          "values": false
        },
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "pluginVersion": "9.5.0",
      "targets": [
        {
          "datasource": {
            "type": "prometheus",
            "uid": "${DS_PROMETHEUS}"
          },
          "editorMode": "code",
          "expr": "sum(rate(http_server_requests_seconds_count{status_code=~\"2..\", service_name=\"${service}\", innovabiz_tenant_id=\"${tenant}\", innovabiz_region_id=\"${region}\"}[$__rate_interval])) / sum(rate(http_server_requests_seconds_count{service_name=\"${service}\", innovabiz_tenant_id=\"${tenant}\", innovabiz_region_id=\"${region}\"}[$__rate_interval])) * 100",
          "instant": false,
          "legendFormat": "Success Rate",
          "range": true,
          "refId": "A"
        }
      ],
      "title": "Taxa de Sucesso",
      "type": "gauge"
    }
  ],
  "refresh": "10s",
  "revision": 1,
  "schemaVersion": 38,
  "style": "dark",
  "tags": ["innovabiz", "${service}", "template"],
  "templating": {
    "list": [
      {
        "current": {
          "selected": false,
          "text": "Prometheus",
          "value": "Prometheus"
        },
        "hide": 0,
        "includeAll": false,
        "multi": false,
        "name": "DS_PROMETHEUS",
        "options": [],
        "query": "prometheus",
        "queryValue": "",
        "refresh": 1,
        "regex": "",
        "skipUrlSync": false,
        "type": "datasource"
      },
      {
        "current": {
          "selected": true,
          "text": "default",
          "value": "default"
        },
        "datasource": {
          "type": "prometheus",
          "uid": "${DS_PROMETHEUS}"
        },
        "definition": "label_values(innovabiz_tenant_id)",
        "hide": 0,
        "includeAll": false,
        "label": "Tenant",
        "multi": false,
        "name": "tenant",
        "options": [],
        "query": {
          "query": "label_values(innovabiz_tenant_id)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "type": "query"
      },
      {
        "current": {
          "selected": true,
          "text": "br",
          "value": "br"
        },
        "datasource": {
          "type": "prometheus",
          "uid": "${DS_PROMETHEUS}"
        },
        "definition": "label_values(innovabiz_region_id)",
        "hide": 0,
        "includeAll": false,
        "label": "Região",
        "multi": false,
        "name": "region",
        "options": [],
        "query": {
          "query": "label_values(innovabiz_region_id)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "type": "query"
      },
      {
        "current": {
          "selected": true,
          "text": "payment-gateway",
          "value": "payment-gateway"
        },
        "datasource": {
          "type": "prometheus",
          "uid": "${DS_PROMETHEUS}"
        },
        "definition": "label_values(service_name)",
        "hide": 0,
        "includeAll": false,
        "label": "Serviço",
        "multi": false,
        "name": "service",
        "options": [],
        "query": {
          "query": "label_values(service_name)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "type": "query"
      },
      {
        "current": {
          "selected": true,
          "text": "production",
          "value": "production"
        },
        "datasource": {
          "type": "prometheus",
          "uid": "${DS_PROMETHEUS}"
        },
        "definition": "label_values(innovabiz_deployment_environment)",
        "hide": 0,
        "includeAll": false,
        "label": "Ambiente",
        "multi": false,
        "name": "environment",
        "options": [],
        "query": {
          "query": "label_values(innovabiz_deployment_environment)",
          "refId": "StandardVariableQuery"
        },
        "refresh": 2,
        "regex": "",
        "skipUrlSync": false,
        "sort": 1,
        "type": "query"
      }
    ]
  },
  "time": {
    "from": "now-3h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": ["5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"],
    "time_options": ["5m", "15m", "1h", "6h", "12h", "24h", "2d", "7d", "30d"]
  },
  "timezone": "browser",
  "title": "INNOVABIZ - ${service} - Template Dashboard",
  "uid": "innovabiz-${service}-template",
  "version": 1,
  "weekStart": "monday",
  "description": "Dashboard template para serviços INNOVABIZ"
}
```

## Dashboards Padrão Requeridos

Cada módulo deve implementar os seguintes dashboards padrão, adaptados às suas métricas específicas:

### 1. Dashboard Operacional

**Objetivo**: Monitoramento diário das operações do serviço  
**Público-alvo**: Operadores, SREs, Desenvolvedores

**Seções e Painéis Recomendados**:
- **Visão Geral**
  - Taxa de requisições (requests/s)
  - Latência (p50, p95, p99)
  - Taxa de erros e código de status
  - Tempo de resposta por endpoint

- **Recursos do Sistema**
  - Uso de CPU
  - Uso de memória
  - Conexões de rede
  - I/O de disco

- **Banco de Dados**
  - Tempo de resposta de consultas
  - Conexões ativas
  - Taxa de cache hit/miss
  - Tamanho das tabelas principais

- **Negócio**
  - Volume de transações por tipo
  - Taxa de aprovação/rejeição
  - Valor médio de transação
  - SLA compliance

### 2. Dashboard de Troubleshooting

**Objetivo**: Diagnóstico aprofundado de problemas  
**Público-alvo**: Desenvolvedores, SREs

**Seções e Painéis Recomendados**:
- **Logs e Traces**
  - Taxa de logs por nível (ERROR, WARN, INFO)
  - Top erros por frequência
  - Distribuição de duração dos traces
  - Spans mais lentos

- **Erros e Exceções**
  - Taxa de exceções por tipo
  - Stack traces mais comuns
  - Dependências com falhas
  - Circuit breakers ativos

- **Correlação**
  - Latência vs. tráfego
  - Erros vs. deployments
  - Erros vs. uso de recursos
  - Heatmap de duração de requisições

### 3. Dashboard Executivo

**Objetivo**: Visão gerencial do serviço  
**Público-alvo**: Gerentes, Líderes Técnicos

**Seções e Painéis Recomendados**:
- **KPIs de Serviço**
  - Disponibilidade (%)
  - Tempo médio entre falhas (MTBF)
  - Tempo médio de recuperação (MTTR)
  - SLAs/SLOs compliance

- **Negócio**
  - Métricas de negócio por tenant/região
  - Tendências de utilização
  - Comparação com períodos anteriores
  - Alertas ativos

### 4. Dashboard de Capacidade

**Objetivo**: Análise de tendências e planejamento  
**Público-alvo**: Arquitetos, SREs, Planejadores

**Seções e Painéis Recomendados**:
- **Tendências de Recursos**
  - Crescimento de CPU/Memória/Disco (4 semanas)
  - Previsão de saturação
  - Análise de sazonalidade
  - Picos de utilização

- **Escalabilidade**
  - Correlação entre réplicas e performance
  - Eficiência de recursos
  - Bottlenecks identificados
  - Recomendações de escala

## Variáveis Padrão

Todos os dashboards devem incluir as seguintes variáveis para filtragem multi-dimensional:

| Variável | Consulta | Descrição |
|----------|----------|-----------|
| `tenant` | `label_values(innovabiz_tenant_id)` | ID do tenant para filtro multi-tenant |
| `region` | `label_values(innovabiz_region_id)` | ID da região para filtro multi-regional |
| `service` | `label_values(service_name)` | Nome do serviço |
| `environment` | `label_values(innovabiz_deployment_environment)` | Ambiente (dev, staging, prod) |
| `module` | `label_values(innovabiz_module_id)` | ID do módulo INNOVABIZ |

## Organização Visual e Hierarquia

O layout dos dashboards deve seguir uma estrutura hierárquica consistente:

1. **Nível 1**: Visão geral e KPIs principais (topo)
2. **Nível 2**: Métricas detalhadas por categoria (meio)
3. **Nível 3**: Detalhes específicos e correlações (parte inferior)

Para consistency visual:

- Use cores padronizadas da paleta INNOVABIZ
- Mantenha unidades consistentes em painéis relacionados
- Forneça links para dashboards relacionados
- Inclua descrições em painéis complexos
- Agrupe painéis em seções lógicas (linhas)

## Melhores Práticas

1. **Consistência de Nomenclatura**
   - Use prefixo `INNOVABIZ - [Módulo] - [Propósito]` para todos os dashboards
   - UIDs consistentes com formato `innovabiz-[módulo]-[propósito]`
   - Mantenha tags padronizadas (`innovabiz`, `[nome-módulo]`, `[tipo-dashboard]`)

2. **Performance**
   - Prefira rate() sobre irate() para estabilidade
   - Use $__rate_interval para consistência
   - Limite a quantidade de séries por painel (<10)
   - Agrupe dados quando apropriado para reduzir cardinosidade

3. **Multi-dimensionalidade**
   - Aplique filtros de tenant/região em todas as consultas
   - Use templating para variáveis multi-contexto
   - Garanta hierarquia visual consistente com contexto INNOVABIZ

4. **Usabilidade**
   - Defina thresholds com cores intuitivas (verde→amarelo→laranja→vermelho)
   - Adicione descrições para painéis complexos
   - Forneça tooltips informativos
   - Inclua links para documentação relevante

5. **Interatividade**
   - Configure drill-down para detalhamento
   - Inclua links para trace/logs relacionados
   - Aplique anotações para eventos importantes (deployments, incidentes)
   - Permita ajuste de thresholds em ambientes não-prod

## Checklist de Validação

- [ ] Variáveis de contexto multi-dimensional implementadas
- [ ] Nomenclatura consistente com padrões INNOVABIZ
- [ ] Todas as consultas filtradas por variáveis de contexto
- [ ] Unidades apropriadas em todos os painéis
- [ ] Thresholds configurados conforme diretrizes
- [ ] Links para dashboards relacionados implementados
- [ ] Anotações para eventos-chave configuradas
- [ ] Painéis organizados em grupos lógicos
- [ ] Dashboard testado com diferentes seleções de variáveis
- [ ] Exportado como JSON para versionamento

## Recursos Adicionais

- [Documentação Grafana](https://grafana.com/docs/)
- [Portal de Observabilidade INNOVABIZ](https://observability.innovabiz.com)
- [Repositório de Dashboards Padrão](https://github.com/innovabiz/observability-dashboards)
- [Biblioteca de Painéis Reutilizáveis](https://grafana.innovabiz.com/library)