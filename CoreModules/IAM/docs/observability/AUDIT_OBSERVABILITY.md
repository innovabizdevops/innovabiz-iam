# Observabilidade do Serviço de Auditoria IAM

## Visão Geral

Este documento descreve a implementação completa de observabilidade para o serviço de auditoria multi-contexto do IAM na plataforma INNOVABIZ. A abordagem adota os princípios de Observabilidade Total, integrando métricas Prometheus, alertas automatizados e dashboards Grafana para garantir visibilidade operacional completa.

## Arquitetura de Observabilidade

A arquitetura de observabilidade do serviço de auditoria IAM foi projetada para atender aos requisitos de um sistema multi-tenant e multi-regional, com foco em:

- **Rastreabilidade completa**: Todas as operações são instrumentadas com identificadores únicos
- **Dimensões de contexto**: Todas as métricas incluem labels para tenant_id e regional_context
- **Granularidade adaptativa**: Métricas disponíveis em diferentes níveis de detalhe
- **Alertas inteligentes**: Regras de alerta com base em anomalias e thresholds dinâmicos
- **Visualização integrada**: Dashboards Grafana pré-configurados para análise operacional

## Componentes Implementados

### 1. Configuração do Prometheus

- **Arquivo**: `config/observability/prometheus.yml`
- **Descrição**: Configuração do Prometheus para scraping de métricas do serviço de auditoria e componentes relacionados
- **Funcionalidades**:
  - Scraping de múltiplos ambientes (prod, staging, qa, dev)
  - Configurações otimizadas para alta disponibilidade
  - Labels automáticos para ambiente e região
  - Integração com PostgreSQL, KrakenD e Redis

### 2. Regras de Alerta

- **Arquivo**: `config/observability/alert_rules.yml`
- **Descrição**: Regras de alerta Prometheus para monitoramento proativo
- **Alertas Configurados**:
  - Disponibilidade do serviço
  - Latência elevada
  - Taxa de erros HTTP
  - Utilização de CPU/memória
  - Saturação de conexões de banco de dados
  - Falhas em políticas de retenção
  - Falhas de conformidade regional### 3. Dashboard Grafana

- **Arquivo**: `config/observability/grafana-dashboards/audit_service_dashboard.json`
- **Descrição**: Dashboard Grafana pré-configurado para visualização das métricas do serviço de auditoria
- **Painéis Incluídos**:
  - Taxa de eventos por minuto
  - Distribuição de eventos por tipo
  - Percentis de latência de processamento (P50, P95)
  - Taxa de erros HTTP
  - Contagem de eventos processados por política de retenção
  - Distribuição de eventos de conformidade por framework
  - Distribuição de eventos por contexto regional
  - Variáveis de template para filtragem por tenant e contexto regional

### 4. Instrumentação de Métricas

- **Diretório**: `src/api/app/metrics/`
- **Arquivos Principais**:
  - `audit_metrics.py`: Definições de métricas e decoradores de instrumentação
  - `__init__.py`: Exportação de funções e componentes
  - `examples/audit_routes_example.py`: Exemplo de integração com FastAPI

#### 4.1 Métricas Implementadas

##### Métricas de Eventos de Auditoria
- `audit_events_total`: Contador de eventos de auditoria processados
- `audit_event_processing_duration`: Histograma de duração do processamento
- `audit_event_size_bytes`: Histograma do tamanho dos eventos em bytes

##### Métricas de Políticas de Retenção
- `audit_retention_policies_active`: Gauge para número de políticas ativas
- `audit_retention_events_processed_total`: Contador de eventos processados
- `audit_retention_policy_execution_duration`: Histograma de duração de execução
- `audit_retention_policy_success`: Contador de execuções bem-sucedidas
- `audit_retention_policy_failure`: Contador de falhas por tipo de erro

##### Métricas de Conformidade
- `audit_compliance_events_total`: Contador de eventos de conformidade
- `audit_compliance_check_duration`: Histograma de duração das verificações
- `audit_regional_compliance_status`: Gauge para status de conformidade regional

##### Métricas HTTP
- `http_requests_total`: Contador de requisições HTTP
- `http_request_duration_seconds`: Histograma de duração das requisições
- `http_response_size_bytes`: Histograma do tamanho das respostas

##### Métricas de Status do Serviço
- `audit_service_info`: Informações sobre o serviço (versão, ambiente)
- `audit_service_uptime_seconds`: Tempo de atividade do serviço
- `audit_service_health_status`: Status de saúde por componente#### 4.2 Decoradores de Instrumentação

A instrumentação do serviço é baseada em decoradores que podem ser aplicados a funções e métodos assíncronos:

##### `instrument_audit_event_processing`
- **Uso**: Decora funções que processam eventos de auditoria
- **Métricas**: Incrementa `audit_events_total` e registra duração/tamanho
- **Contexto**: Extrai automaticamente labels de tenant, região e severidade
- **Exemplo**:
  ```python
  @instrument_audit_event_processing
  async def process_audit_event(event):
      # Processamento do evento
      return result
  ```

##### `instrument_retention_policy`
- **Uso**: Decora funções que executam políticas de retenção
- **Métricas**: Registra sucesso/falha, duração e contagem de eventos processados
- **Contexto**: Extrai automaticamente labels de tenant, região e tipo de política
- **Exemplo**:
  ```python
  @instrument_retention_policy
  async def apply_retention_policy(policy_type, tenant_id, regional_context, **kwargs):
      # Aplicação da política
      return result
  ```

##### `instrument_compliance_check`
- **Uso**: Decora funções que verificam conformidade regulatória
- **Métricas**: Registra resultados de conformidade e duração das verificações
- **Contexto**: Extrai automaticamente labels de tenant, região, framework e regulação
- **Exemplo**:
  ```python
  @instrument_compliance_check
  async def verify_compliance(tenant_id, regional_context, framework, regulation, **kwargs):
      # Verificação de conformidade
      return {"compliant": True, "details": {...}}
  ```

#### 4.3 Middleware HTTP

O middleware HTTP automaticamente instrumenta todas as requisições:

```python
# Registrar middleware na aplicação FastAPI
app.middleware("http")(metrics_middleware)
```

- **Funcionalidades**:
  - Contabilização de requisições por endpoint e método
  - Medição de latência de resposta
  - Captura de tamanho de resposta
  - Propagação de contexto multi-tenant e multi-regional
  - Registro de status code das respostas

#### 4.4 Inicialização de Métricas

Para inicializar as métricas em uma aplicação FastAPI:

```python
from src.api.app.metrics import init_metrics, setup_service_info

# Criar aplicativo
app = FastAPI(...)

# Inicializar métricas
init_metrics(app)  # Configura o endpoint /metrics e middleware

# Configurar informações do serviço
setup_service_info(
    version="1.0.0",
    build_id="build-123",
    commit_hash="abc123",
    environment="production",
    region="global"
)
```## Integração com Serviços Existentes

### 1. Integração com KrakenD API Gateway

O serviço de auditoria expõe suas métricas para serem coletadas pelo KrakenD API Gateway, permitindo uma visão unificada de toda a stack de autenticação e autorização:

```yaml
# Configuração no KrakenD
telemetry:
  prometheus:
    listen_address: ":9090"
    collection_time: "60s"
    remote_services:
      - name: "iam-audit"
        url: "http://iam-audit-service:8000/metrics"
        service_name: "iam_audit"
        interval: "15s"
```

### 2. Integração Multi-Contexto

As métricas são dimensionadas para suportar múltiplos contextos:

- **Multi-tenant**: Todas as métricas incluem o label `tenant_id`
- **Multi-regional**: Todas as métricas incluem o label `regional_context`
- **Multi-ambiente**: A configuração de scraping Prometheus separa os ambientes
- **Multi-compliance**: Métricas de conformidade são separadas por `framework` e `regulation`

Esta abordagem permite:

- Análises por tenant específico
- Comparações entre regiões regulatórias
- Isolamento de problemas por contexto
- Dashboards filtráveis por qualquer dimensão contextual

### 3. Integração com Serviços de Observabilidade

#### Prometheus Federation

O serviço suporta federação Prometheus, permitindo que instâncias hierárquicas colete métricas:

```yaml
# Exemplo de configuração de federação
scrape_configs:
  - job_name: 'federate-iam-audit'
    scrape_interval: 15s
    honor_labels: true
    metrics_path: '/federate'
    params:
      'match[]':
        - '{job="iam-audit"}'
    static_configs:
      - targets:
        - 'prometheus-iam:9090'
```

#### Alertmanager Integration

Os alertas definidos em `alert_rules.yml` são enviados para o Alertmanager central, que pode distribuir notificações por múltiplos canais:

- Slack/Teams para notificações imediatas
- Email para relatórios diários
- PagerDuty para alertas críticos
- Webhook para integração com sistemas de ticket

#### Grafana Provisioning

O dashboard Grafana é provisionado automaticamente usando a API de provisioning:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d @config/observability/grafana-dashboards/audit_service_dashboard.json \
  http://grafana:3000/api/dashboards/db
```## Boas Práticas de Observabilidade

### 1. Convenções de Nomenclatura

Todas as métricas seguem convenções de nomenclatura consistentes:

- **Prefixo de domínio**: `audit_` para métricas específicas do serviço de auditoria
- **Sufixo de tipo**: `_total` para contadores, `_seconds` para durações, `_bytes` para tamanhos
- **Nomes descritivos**: Nomes que descrevem claramente o que está sendo medido
- **Labels consistentes**: Conjunto consistente de labels em métricas relacionadas

### 2. Dimensões de Métricas

As seguintes dimensões são aplicadas consistentemente em todas as métricas:

- **tenant_id**: Identificador do tenant (obrigatório)
- **regional_context**: Contexto regional (obrigatório)
- **event_type**: Tipo de evento para métricas de eventos
- **severity**: Severidade para métricas de eventos
- **policy_type**: Tipo de política para métricas de retenção
- **framework**: Framework regulatório para métricas de conformidade
- **regulation**: Regulação específica para métricas de conformidade

### 3. Governança de Observabilidade

- **Auditoria de métricas**: Revisão trimestral de métricas para garantir relevância
- **Otimização de cardinality**: Controle de cardinalidade para evitar explosão de séries temporais
- **Políticas de retenção**: Políticas de retenção de métricas alinhadas com requisitos regulatórios
- **Níveis de acesso**: Controle de acesso RBAC para dashboards por tenant e região

## Troubleshooting

### Problemas Comuns e Soluções

#### 1. Alta cardinalidade de métricas

**Sintoma**: Prometheus apresenta desempenho lento ou erros de memória.

**Causa**: Excesso de valores únicos em labels como tenant_id.

**Solução**:
- Verificar uso de `record rules` para agregar métricas por grupos
- Considerar uso de cluster Thanos para séries temporais longas
- Implementar `exemplars` para rastreabilidade seletiva

#### 2. Latência na coleta de métricas

**Sintoma**: Dashboards Grafana mostram dados desatualizados.

**Causa**: Intervalo de scraping muito longo ou timeout nos endpoints de métricas.

**Solução**:
- Ajustar `scrape_interval` no Prometheus
- Verificar performance do endpoint `/metrics`
- Implementar caching de métricas custosas

#### 3. Alertas com falsos positivos

**Sintoma**: Alertas disparando sem problema real.

**Causa**: Thresholds muito sensíveis ou não específicos por contexto.

**Solução**:
- Implementar alertas com base em múltiplas condições
- Configurar thresholds dinâmicos por tenant ou região
- Adicionar regras de silenciamento para manutenções planejadas## Próximos Passos

### 1. Extensão de Observabilidade

Para a próxima fase de desenvolvimento, recomenda-se implementar as seguintes extensões:

- **Tracing distribuído**: Integração com OpenTelemetry para rastreamento de requisições entre serviços
- **Logging estruturado**: Integração com ELK Stack para correlação entre logs e métricas
- **Métricas de negócio**: Implementação de métricas orientadas a KPIs de negócio
- **Machine Learning para detecção de anomalias**: Modelo para identificar comportamentos anômalos em padrões de auditoria

### 2. Automação de Observabilidade

- **Auto-scaling baseado em métricas**: Escalar o serviço com base em métricas de carga
- **Testes de caos**: Validar resiliência e observabilidade sob condições de falha
- **Automação de runbooks**: Implementar recuperação automática para cenários de falha conhecidos
- **Observabilidade como código**: Gerenciar toda configuração de observabilidade via GitOps

### 3. Integração com IAM Advanced Analytics

- **Análise preditiva**: Identificação precoce de possíveis falhas de conformidade
- **Painel executivo**: Dashboard consolidado para C-level com métricas de conformidade
- **Relatórios automatizados**: Geração programada de relatórios de auditoria por região e framework

## Conclusão

A implementação da observabilidade para o serviço de auditoria do IAM INNOVABIZ estabelece um novo padrão para monitoramento e operação de sistemas multi-contexto. A abordagem adotada permite:

1. **Visibilidade total**: Monitoramento completo de todas as operações e eventos do serviço
2. **Detecção proativa**: Alertas inteligentes que identificam problemas antes de afetarem usuários
3. **Conformidade verificável**: Métricas de conformidade que demonstram aderência a frameworks regulatórios
4. **Escalabilidade operacional**: Instrumentação eficiente que suporta crescimento sem degradar performance

A arquitetura de observabilidade implementada é flexível o suficiente para evoluir com o sistema, permitindo adicionar novas dimensões de contexto, métricas de negócio e integrações com ferramentas avançadas de análise.

## Referências

1. [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)
2. [Grafana Dashboard Design](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/)
3. [OpenTelemetry for Python](https://opentelemetry.io/docs/instrumentation/python/)
4. [SRE Books - Google](https://sre.google/books/)
5. [INNOVABIZ Platform Standards](http://internal-docs.innovabiz.com/platform/standards/)
6. [Multi-Context Observability Pattern](http://internal-docs.innovabiz.com/platform/patterns/multi-context-observability/)

---

**Documento preparado por:** Equipe de Observabilidade INNOVABIZ  
**Versão:** 1.0.0  
**Data:** 31 de Julho de 2025  
**Status:** Aprovado para Implementação