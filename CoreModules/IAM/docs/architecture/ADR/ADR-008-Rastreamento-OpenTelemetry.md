# ADR-008: Integração de Rastreamento Distribuído com OpenTelemetry para IAM Audit Service

## Status

Aprovado

## Data

2025-07-31

## Contexto

O IAM Audit Service opera em um ambiente distribuído complexo, com interações entre múltiplos serviços, dependências externas, e contextos multi-tenant/multi-região. A observabilidade através de métricas e logs é essencial, mas insuficiente para:

- Compreender o fluxo completo de requisições entre serviços
- Identificar gargalos de performance em chamadas distribuídas
- Correlacionar eventos de auditoria com ações do usuário final
- Detectar falhas em serviços dependentes que impactam o fluxo de auditoria
- Medir latências end-to-end para eventos de auditoria críticos
- Atender requisitos de compliance para visibilidade completa de operações sensíveis

É necessário implementar um sistema de rastreamento distribuído (distributed tracing) que permita visualizar e analisar o fluxo completo de execução entre componentes, com suporte para contextos multi-tenant e multi-região.

## Decisão

Implementar integração com **OpenTelemetry** para rastreamento distribuído no IAM Audit Service, com as seguintes características:

### 1. Instrumentação com OpenTelemetry

#### 1.1. Inicialização do Tracer

```python
from opentelemetry import trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

class TracingIntegration:
    def __init__(self, config: TracingConfig = None):
        self.config = config or TracingConfig()
        
        # Configuração do recurso OpenTelemetry
        resource = Resource.create({
            "service.name": self.config.service_name,
            "service.namespace": self.config.namespace,
            "service.version": self.config.version,
            "deployment.environment": self.config.environment
        })
        
        # Configuração do TracerProvider
        provider = TracerProvider(resource=resource)
        
        # Configuração do exportador OTLP
        otlp_exporter = OTLPSpanExporter(
            endpoint=self.config.otlp_endpoint,
            headers=self.config.otlp_headers
        )
        
        # Processor para envio batch de spans
        span_processor = BatchSpanProcessor(otlp_exporter)
        provider.add_span_processor(span_processor)
        
        # Registro global do provider
        trace.set_tracer_provider(provider)
        
        # Tracer para o serviço
        self.tracer = trace.get_tracer(
            self.config.service_name,
            self.config.version
        )
    
    def instrument_app(self, app: FastAPI):
        """
        Instrumenta automaticamente uma aplicação FastAPI.
        """
        # Instrumentação automática do FastAPI
        FastAPIInstrumentor.instrument_app(
            app,
            tracer_provider=trace.get_tracer_provider(),
            excluded_urls=self.config.excluded_urls,
        )
        
        # Middleware para enriquecimento de contexto multi-tenant e multi-região
        @app.middleware("http")
        async def add_tenant_region_to_spans(request: Request, call_next):
            current_span = trace.get_current_span()
            tenant = request.headers.get("X-Tenant-ID", "default")
            region = request.headers.get("X-Region", "global")
            
            # Adiciona tags de tenant e região ao span atual
            current_span.set_attribute("tenant.id", tenant)
            current_span.set_attribute("region.name", region)
            
            # Processa a requisição normalmente
            response = await call_next(request)
            return response
```

#### 1.2. Configuração de Rastreamento

```python
class TracingConfig(BaseSettings):
    """Configuração para rastreamento distribuído com OpenTelemetry."""
    
    # Configuração básica
    service_name: str = "iam-audit-service"
    namespace: str = "innovabiz"
    version: str = "1.0.0"
    environment: str = "production"
    
    # Configuração do exportador OTLP
    otlp_endpoint: str = "http://otel-collector:4317"
    otlp_headers: Dict[str, str] = {}
    
    # URLs a serem excluídas da instrumentação automática
    excluded_urls: List[str] = ["/health", "/live", "/metrics"]
    
    # Sampling configuration
    sample_ratio: float = 1.0  # 100% por padrão
    
    # Atributos padrão
    default_attributes: Dict[str, str] = {}
    
    # Configurações específicas para diferentes tipos de eventos
    span_configs: Dict[str, Dict] = {
        "audit_event": {
            "sample_ratio": 1.0,  # Sempre amostra eventos de auditoria
            "attributes": ["event_type", "user_id", "resource_id"]
        },
        "compliance_check": {
            "sample_ratio": 1.0,
            "attributes": ["compliance_type", "check_id"]
        },
        "retention_operation": {
            "sample_ratio": 0.25,  # Amostragem reduzida para operações frequentes
            "attributes": ["retention_policy", "affected_records"]
        }
    }
    
    class Config:
        env_prefix = "TRACING_"
        env_nested_delimiter = "__"
```

### 2. Instrumentação de Código de Negócios

#### 2.1. Decorador para Instrumentação de Funções

```python
def traced(span_name: str = None, span_type: str = None, attributes: Dict = None):
    """
    Decorador para instrumentar funções com OpenTelemetry.
    
    Args:
        span_name: Nome personalizado para o span.
        span_type: Tipo de span (audit_event, compliance_check, retention_operation, etc.)
        attributes: Atributos adicionais para o span.
    """
    def decorator(func):
        @wraps(func)
        async def async_wrapper(self, *args, **kwargs):
            # Determina o nome do span
            name = span_name or func.__name__
            
            # Obtém o tracer
            tracer = trace.get_tracer(self.config.service_name, self.config.version)
            
            # Extrai informações de contexto
            tenant = getattr(request.state, "tenant", "default")
            region = getattr(request.state, "region", "global")
            
            # Prepara atributos
            span_attributes = {
                "tenant.id": tenant,
                "region.name": region,
                "span.type": span_type or "function"
            }
            
            # Adiciona atributos específicos da função
            if attributes:
                span_attributes.update(attributes)
                
            # Adiciona atributos de argumentos da função configurados
            if span_type and span_type in self.config.span_configs:
                span_config = self.config.span_configs[span_type]
                for attr in span_config.get("attributes", []):
                    if attr in kwargs:
                        span_attributes[attr] = kwargs[attr]
            
            # Cria o span
            with tracer.start_as_current_span(name, attributes=span_attributes) as span:
                try:
                    # Executa a função original
                    result = await func(self, *args, **kwargs)
                    return result
                except Exception as e:
                    # Marca o span como com erro
                    span.set_status(Status(StatusCode.ERROR))
                    span.record_exception(e)
                    raise
        
        @wraps(func)
        def sync_wrapper(self, *args, **kwargs):
            # Implementação similar para funções síncronas
            # ...
        
        if asyncio.iscoroutinefunction(func):
            return async_wrapper
        return sync_wrapper
    
    return decorator
```

#### 2.2. Exemplos de Uso

```python
class AuditService:
    def __init__(self, config: Config, db: Database, tracing: TracingIntegration):
        self.config = config
        self.db = db
        self.tracing = tracing
    
    @traced(span_type="audit_event")
    async def create_audit_event(self, event: AuditEventCreate, tenant: str, region: str) -> AuditEvent:
        """
        Cria um novo evento de auditoria.
        Automaticamente instrumentado com span de rastreamento.
        """
        # Lógica para criação do evento...
        result = await self.db.insert_audit_event(event, tenant, region)
        return result
    
    @traced(span_type="compliance_check")
    async def verify_compliance(self, check_type: str, resource_id: str, tenant: str, region: str) -> ComplianceResult:
        """
        Verifica compliance para um recurso.
        Automaticamente instrumentado com span de rastreamento.
        """
        # Inicia span manual para uma etapa crítica
        with self.tracing.tracer.start_as_current_span(
            "fetch_compliance_rules",
            attributes={
                "tenant.id": tenant,
                "region.name": region,
                "check_type": check_type
            }
        ) as span:
            rules = await self.fetch_compliance_rules(check_type, tenant, region)
        
        # Lógica para verificação...
        result = await self._evaluate_compliance(rules, resource_id)
        return result
    
    @traced(span_type="retention_operation")
    async def execute_retention_policy(self, policy_name: str, tenant: str, region: str) -> RetentionResult:
        """
        Executa uma política de retenção.
        Automaticamente instrumentado com span de rastreamento.
        """
        # Lógica para execução da política...
        result = await self._apply_retention_rules(policy_name, tenant, region)
        return result
```

### 3. Propagação de Contexto

Implementação para garantir a propagação de contexto de rastreamento entre serviços:

```python
from opentelemetry.context.propagation import get_global_textmap, set_global_textmap
from opentelemetry.propagate import extract, inject
from opentelemetry.propagators.b3 import B3MultiFormat
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

# Configurar propagadores para suportar múltiplos formatos
propagator = CompositePropagator([
    TraceContextTextMapPropagator(),  # W3C Trace Context
    B3MultiFormat(),                  # Zipkin B3
])
set_global_textmap(propagator)

class HttpClientWithTracing:
    """Cliente HTTP com propagação automática de contexto de rastreamento."""
    
    def __init__(self, base_url: str, tracing: TracingIntegration):
        self.base_url = base_url
        self.tracing = tracing
        self.client = httpx.AsyncClient(base_url=base_url)
    
    async def request(
        self,
        method: str,
        url: str,
        tenant: str,
        region: str,
        headers: Dict[str, str] = None,
        **kwargs
    ) -> httpx.Response:
        """
        Realiza requisição HTTP com propagação de contexto de rastreamento.
        """
        # Prepara headers
        headers = headers or {}
        
        # Adiciona headers de contexto
        headers["X-Tenant-ID"] = tenant
        headers["X-Region"] = region
        
        # Injeta contexto de rastreamento nos headers
        inject(headers)
        
        # Cria span para a requisição
        with self.tracing.tracer.start_as_current_span(
            f"HTTP {method}",
            attributes={
                "http.method": method,
                "http.url": url,
                "tenant.id": tenant,
                "region.name": region
            }
        ) as span:
            # Realiza a requisição
            response = await self.client.request(
                method=method,
                url=url,
                headers=headers,
                **kwargs
            )
            
            # Adiciona atributos de resposta ao span
            span.set_attribute("http.status_code", response.status_code)
            
            if response.status_code >= 400:
                span.set_status(Status(StatusCode.ERROR))
            
            return response
```

### 4. Visualização e Análise

Integração com **Jaeger** para visualização e análise de traces:

```yaml
# Docker Compose para ambiente de observabilidade
version: '3.8'
services:
  # Jaeger para visualização de traces
  jaeger:
    image: jaegertracing/all-in-one:1.39
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # Ingestão via HTTP
      - "4317:4317"    # Ingestão via gRPC OTLP
      - "4318:4318"    # Ingestão via HTTP OTLP
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - COLLECTOR_ZIPKIN_HOST_PORT=:9411
      - SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json
    volumes:
      - ./config/jaeger/sampling_strategies.json:/etc/jaeger/sampling_strategies.json

  # OpenTelemetry Collector
  otel-collector:
    image: otel/opentelemetry-collector-contrib:0.70.0
    volumes:
      - ./config/otel-collector/config.yaml:/etc/otel-collector/config.yaml
    command: ["--config=/etc/otel-collector/config.yaml"]
    ports:
      - "4317:4317"  # OTLP gRPC
      - "4318:4318"  # OTLP HTTP
      - "8888:8888"  # Metrics
    depends_on:
      - jaeger
```

Configuração do OpenTelemetry Collector:

```yaml
# config/otel-collector/config.yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
    
  attributes:
    actions:
      - key: tenant.id
        action: insert
        value: "unknown"
      - key: region.name
        action: insert
        value: "unknown"
  
  resource:
    attributes:
      - key: service.namespace
        action: upsert
        value: "innovabiz"

exporters:
  logging:
    loglevel: info
  
  otlp/jaeger:
    endpoint: jaeger:4317
    tls:
      insecure: true

  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: innovabiz

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [attributes, resource, batch]
      exporters: [otlp/jaeger, logging]
    
    metrics:
      receivers: [otlp]
      processors: [attributes, resource, batch]
      exporters: [prometheus, logging]
```

### 5. Integração com Sistema de Observabilidade Existente

Integração com `ObservabilityIntegration` para solução completa:

```python
class ObservabilityIntegration:
    def __init__(
        self,
        metrics_config: MetricsConfig = None,
        tracing_config: TracingConfig = None
    ):
        self.metrics_config = metrics_config or MetricsConfig()
        self.tracing_config = tracing_config or TracingConfig()
        
        # Inicializa métricas
        self.metrics = self._setup_metrics()
        
        # Inicializa tracing
        self.tracing = TracingIntegration(self.tracing_config)
        
        # Router para endpoints de observabilidade
        self.router = APIRouter(tags=["Observability"])
        self._setup_endpoints()
    
    def instrument_app(self, app: FastAPI):
        """
        Instrumenta uma aplicação FastAPI com métricas e tracing.
        """
        # Adiciona middleware de métricas
        app.add_middleware(HTTPMetricsMiddleware, exclude_paths=["/metrics", "/health"])
        
        # Instrumenta app com OpenTelemetry
        self.tracing.instrument_app(app)
        
        # Adiciona endpoints de observabilidade
        app.include_router(self.router)
        
        # Registra handlers para ciclo de vida
        @app.on_event("startup")
        async def startup_event():
            # Inicialização de recursos
            pass
        
        @app.on_event("shutdown")
        async def shutdown_event():
            # Limpeza de recursos
            pass
```

## Alternativas Consideradas

### 1. Implementação Própria de Rastreamento

**Prós:**
- Controle total sobre implementação
- Otimizado especificamente para o IAM Audit Service
- Sem dependências externas

**Contras:**
- Alto custo de desenvolvimento
- Falta de interoperabilidade com outros sistemas
- Necessidade de manutenção própria
- Ausência de ferramentas de visualização maduras

### 2. Uso de Zipkin ou Jaeger Diretamente

**Prós:**
- Ferramentas maduras com comunidade ativa
- Visualização integrada
- Documentação extensa

**Contras:**
- Menos flexibilidade para futuro
- Potencial vendor lock-in
- Foco específico em tracing, sem solução completa de observabilidade

### 3. Solução SaaS Comercial (Datadog, NewRelic, Dynatrace)

**Prós:**
- Solução completa integrada
- Menor esforço de configuração e manutenção
- Funcionalidades avançadas de análise e AI/ML

**Contras:**
- Custos operacionais contínuos
- Dependência de fornecedor externo
- Potencial preocupação com dados sensíveis de auditoria
- Complexidade para configuração multi-tenant

## Consequências

### Positivas

- **Visibilidade end-to-end**: Compreensão completa do fluxo de execução
- **Troubleshooting facilitado**: Rápida identificação de gargalos e falhas
- **Correlação entre serviços**: Visualização de dependências entre componentes
- **Medição precisa de latências**: Análise de performance distribuída
- **Flexibilidade**: Padrão aberto com múltiplas opções de backend
- **Integração multi-contexto**: Suporte nativo para tenant e região
- **Conformidade regulatória**: Capacidade de demonstrar fluxo completo de operações sensíveis

### Negativas

- **Overhead de performance**: Impacto mínimo devido ao rastreamento
- **Complexidade adicional**: Necessidade de configuração e manutenção
- **Volume de dados**: Potencial grande volume de dados de telemetria
- **Necessidade de filtragem**: Requisito de sampling estratégico para balancear detalhe e volume

### Mitigação de Riscos

- Implementar políticas de sampling inteligente
- Configurar retenção adequada para dados de traces
- Criar dashboards específicos para análise de performance
- Treinar equipe em análise de traces distribuídos
- Automatizar alertas baseados em padrões anômalos de traces
- Estabelecer processos para revisão periódica de thresholds

## Conformidade com Padrões

- **OpenTelemetry Specification**: Aderência ao padrão aberto
- **W3C Trace Context**: Padrão para propagação de contexto
- **ISO/IEC 25010**: Atributos de qualidade para manutenibilidade e testabilidade
- **PCI DSS 4.0**: Requisitos de rastreabilidade de operações (10.2)
- **GDPR/LGPD**: Capacidade de demonstrar fluxo completo de processamento de dados
- **INNOVABIZ Platform Observability Standards v2.5**

## Implementação

A implementação inclui:

1. **Módulo `tracing.integration`**:
   - Classe `TracingIntegration` para inicialização e configuração
   - Integração com FastAPI
   - Propagação de contexto

2. **Módulo `tracing.decorators`**:
   - Decoradores para instrumentação de funções
   - Utilitários para criação de spans customizados
   - Helpers para extração de atributos

3. **Módulo `tracing.clients`**:
   - Clientes HTTP e outros com propagação automática de contexto
   - Instrumentação de operações de banco de dados

4. **Configuração de Infraestrutura**:
   - Configuração do OpenTelemetry Collector
   - Configuração do Jaeger
   - Configuração de sampling

## Exemplos de Visualização

### 1. Trace de Processamento de Evento de Auditoria

```
[Trace] Processamento de Evento de Auditoria (duração total: 153ms)
├─ [Span] POST /api/v1/audit/events (125ms)
│  ├─ [Span] validate_audit_event (12ms)
│  ├─ [Span] authenticate_request (28ms)
│  │  └─ [Span] HTTP GET auth-service/validate-token (24ms)
│  ├─ [Span] create_audit_event (67ms)
│  │  ├─ [Span] fetch_compliance_rules (15ms)
│  │  │  └─ [Span] HTTP GET compliance-service/rules (12ms)
│  │  ├─ [Span] check_compliance (18ms)
│  │  ├─ [Span] persist_event (25ms)
│  │  │  └─ [Span] DB INSERT audit_events (22ms)
│  │  └─ [Span] publish_event_notification (7ms)
│  │     └─ [Span] Kafka SEND audit-events-topic (5ms)
│  └─ [Span] format_response (5ms)
└─ [Span] Background retention check (28ms) 
```

### 2. Trace de Execução de Política de Retenção

```
[Trace] Execução de Política de Retenção GDPR (duração total: 1.25s)
├─ [Span] scheduled_retention_job (1.25s)
│  ├─ [Span] load_retention_policy (45ms)
│  │  └─ [Span] DB SELECT retention_policies (42ms)
│  ├─ [Span] identify_records_for_deletion (320ms)
│  │  └─ [Span] DB QUERY audit_events (315ms)
│  ├─ [Span] backup_records (450ms)
│  │  ├─ [Span] format_backup (125ms)
│  │  └─ [Span] S3 PUT backup-file (320ms)
│  ├─ [Span] delete_records (380ms)
│  │  └─ [Span] DB DELETE audit_events (375ms)
│  └─ [Span] record_retention_execution (52ms)
│     └─ [Span] DB INSERT retention_executions (48ms)
```

## Referências

1. OpenTelemetry Documentation - https://opentelemetry.io/docs/
2. W3C Trace Context Specification - https://www.w3.org/TR/trace-context/
3. Jaeger Tracing - https://www.jaegertracing.io/docs/
4. Distributed Tracing: A Complete Guide - https://lightstep.com/distributed-tracing
5. FastAPI Observability Best Practices - https://fastapi.tiangolo.com/advanced/opentelemetry/
6. INNOVABIZ Platform Observability Standards v2.5 (Internal Document)
7. Multi-tenant Tracing Patterns - https://www.splunk.com/en_us/blog/devops/multi-tenant-distributed-tracing.html