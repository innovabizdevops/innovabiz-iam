# Template para Instrumentação OpenTelemetry - Python (FastAPI)

## Visão Geral

Este documento fornece o template padrão INNOVABIZ para instrumentação de aplicações Python (FastAPI) com OpenTelemetry, garantindo consistência nas métricas, logs e traces coletados em toda a plataforma.

## Pré-requisitos

- Python 3.8 ou superior
- FastAPI 0.95.0 ou superior
- Acesso ao OpenTelemetry Collector (via variáveis de ambiente)

## Instalação das Dependências

```bash
# Instalar OpenTelemetry Core
pip install opentelemetry-api opentelemetry-sdk

# Instalar exportadores
pip install opentelemetry-exporter-otlp

# Instalar instrumentações automáticas
pip install opentelemetry-instrumentation-fastapi opentelemetry-instrumentation-requests
pip install opentelemetry-instrumentation-sqlalchemy opentelemetry-instrumentation-redis
pip install opentelemetry-instrumentation-logging opentelemetry-instrumentation-kafka

# Instalar SDK INNOVABIZ (opcional, mas recomendado)
pip install innovabiz-observability-sdk
```

## Template de Implementação

### 1. Arquivo de Configuração - `app/observability/telemetry.py`

```python
# Configuração do OpenTelemetry para Python (FastAPI)
import os
import logging
from typing import Dict, Any, Optional, Callable

from opentelemetry import trace, metrics, context
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentation
from opentelemetry.instrumentation.requests import RequestsInstrumentation
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentation
from opentelemetry.instrumentation.redis import RedisInstrumentation
from opentelemetry.instrumentation.logging import LoggingInstrumentation
from opentelemetry.instrumentation.kafka import KafkaInstrumentation
from opentelemetry.propagate import set_global_textmap
from opentelemetry.propagators.composite import CompositePropagator
from opentelemetry.propagators.b3 import B3Format
from opentelemetry.propagators.jaeger import JaegerPropagator
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

logger = logging.getLogger(__name__)

# Tente importar o propagador personalizado INNOVABIZ, ou use implementação fallback
try:
    from innovabiz_observability_sdk.propagation import InnovabizContextPropagator
except ImportError:
    # Implementação fallback do propagador de contexto INNOVABIZ
    class InnovabizContextPropagator:
        def inject(self, carrier, context=None):
            tenant_id = os.getenv("TENANT_ID", "default")
            region_id = os.getenv("REGION_ID", "default")
            
            carrier["x-innovabiz-tenant-id"] = tenant_id
            carrier["x-innovabiz-region-id"] = region_id
            carrier["x-innovabiz-context-version"] = "1.0"
            
            return carrier
        
        def extract(self, carrier, context=None):
            # Em uma implementação completa, extrairíamos informações do carrier
            # e as anexaríamos ao contexto
            return context or {}
        
        def fields(self):
            return ["x-innovabiz-tenant-id", "x-innovabiz-region-id", "x-innovabiz-context-version"]


def setup_opentelemetry(service_name: str, module_id: str, service_version: str) -> None:
    """
    Configura OpenTelemetry para a aplicação
    
    Args:
        service_name: Nome do serviço
        module_id: ID do módulo INNOVABIZ
        service_version: Versão do serviço
    """
    # Informações multi-contexto INNOVABIZ
    resource = Resource.create({
        ResourceAttributes.SERVICE_NAME: service_name,
        ResourceAttributes.SERVICE_VERSION: service_version,
        "innovabiz.module.id": module_id,
        "innovabiz.deployment.environment": os.getenv("ENVIRONMENT", "development"),
        "innovabiz.tenant.id": os.getenv("TENANT_ID", "default"),
        "innovabiz.region.id": os.getenv("REGION_ID", "default"),
    })

    # Configuração do propagador composto
    set_global_textmap(CompositePropagator([
        InnovabizContextPropagator(),  # Propagador personalizado INNOVABIZ
        TraceContextTextMapPropagator(),
        B3Format(),
        JaegerPropagator(),
    ]))

    # Configuração do tracer provider
    tracer_provider = TracerProvider(resource=resource)
    
    # Headers para exportadores
    headers = {
        "x-innovabiz-tenant-id": os.getenv("TENANT_ID", "default"),
        "x-innovabiz-region-id": os.getenv("REGION_ID", "default"),
    }
    
    # Exportador OTLP para traces
    otlp_trace_exporter = OTLPSpanExporter(
        endpoint=os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
        headers=headers
    )
    
    # Adiciona processador de spans
    tracer_provider.add_span_processor(BatchSpanProcessor(otlp_trace_exporter))
    
    # Se em ambiente de desenvolvimento, adiciona console exporter para debugging
    if os.getenv("ENVIRONMENT") == "development":
        tracer_provider.add_span_processor(BatchSpanProcessor(ConsoleSpanExporter()))
    
    # Define o tracer provider global
    trace.set_tracer_provider(tracer_provider)
    
    # Exportador OTLP para métricas
    otlp_metric_exporter = OTLPMetricExporter(
        endpoint=os.getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4317"),
        headers=headers
    )
    
    # Configuração do leitor de métricas
    metric_reader = PeriodicExportingMetricReader(
        exporter=otlp_metric_exporter,
        export_interval_millis=15000
    )
    
    # Configuração do meter provider
    metrics.set_meter_provider(MeterProvider(resource=resource, metric_readers=[metric_reader]))
    
    # Instrumentação automática
    FastAPIInstrumentation().instrument()
    RequestsInstrumentation().instrument()
    SQLAlchemyInstrumentation().instrument()
    RedisInstrumentation().instrument()
    LoggingInstrumentation().instrument()
    KafkaInstrumentation().instrument()
    
    logger.info(f"OpenTelemetry inicializado para {service_name} v{service_version}")


def create_custom_metrics():
    """
    Cria e registra métricas customizadas para o serviço
    """
    meter = metrics.get_meter("innovabiz-custom-metrics")
    
    # Counter para transações processadas
    transaction_counter = meter.create_counter(
        name="transactions.count",
        description="Contador de transações processadas",
        unit="1",
    )
    
    # Histogram para latência de transações
    transaction_duration = meter.create_histogram(
        name="transaction.duration",
        description="Duração das transações",
        unit="ms",
    )
    
    # Up/Down Counter para usuários ativos
    active_users = meter.create_up_down_counter(
        name="users.active",
        description="Usuários ativos no momento",
        unit="1",
    )
    
    # Observer para métricas do sistema
    def _get_memory_usage(observer):
        """Callback para observar uso de memória"""
        import psutil
        memory = psutil.virtual_memory()
        observer.observe(memory.used, {"innovabiz.resource.type": "memory"})
    
    system_memory = meter.create_observable_gauge(
        name="system.memory.usage",
        description="Uso de memória",
        unit="bytes",
        callbacks=[_get_memory_usage]
    )
    
    return {
        "transaction_counter": transaction_counter,
        "transaction_duration": transaction_duration,
        "active_users": active_users,
        "system_memory": system_memory
    }


# Middleware para contexto multi-dimensional
class InnovabizContextMiddleware:
    """Middleware para gerenciar contexto multi-dimensional do INNOVABIZ."""
    
    def __init__(
        self, 
        app, 
        tenant_header: str = "x-tenant-id", 
        region_header: str = "x-region-id",
        default_tenant: str = "default",
        default_region: str = "default"
    ):
        self.app = app
        self.tenant_header = tenant_header
        self.region_header = region_header
        self.default_tenant = default_tenant
        self.default_region = default_region
    
    async def __call__(self, scope, receive, send):
        if scope["type"] != "http":
            return await self.app(scope, receive, send)
            
        # Extrair valores de cabeçalho
        headers = dict(scope.get("headers", []))
        
        tenant_id_bytes = headers.get(self.tenant_header.lower().encode(), None)
        region_id_bytes = headers.get(self.region_header.lower().encode(), None)
        
        tenant_id = tenant_id_bytes.decode() if tenant_id_bytes else (os.getenv("TENANT_ID") or self.default_tenant)
        region_id = region_id_bytes.decode() if region_id_bytes else (os.getenv("REGION_ID") or self.default_region)
        
        # Injetar no escopo para acesso downstream
        scope["innovabiz_context"] = {
            "tenant_id": tenant_id,
            "region_id": region_id
        }
        
        # Modificar a função send para injetar cabeçalhos de resposta
        original_send = send
        
        async def send_wrapper(message):
            if message["type"] == "http.response.start":
                # Adicionar cabeçalhos de contexto à resposta
                headers = message.get("headers", [])
                headers.append((b"x-innovabiz-tenant-id", tenant_id.encode()))
                headers.append((b"x-innovabiz-region-id", region_id.encode()))
                message["headers"] = headers
            
            await original_send(message)
            
        return await self.app(scope, receive, send_wrapper)
```

### 2. Integração no Arquivo Principal - `app/main.py`

```python
import os
import time
from typing import Dict, Any

from fastapi import FastAPI, Request, Response, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

# Importar configuração de telemetria
from app.observability.telemetry import setup_opentelemetry, create_custom_metrics, InnovabizContextMiddleware

# Inicializar OpenTelemetry
setup_opentelemetry(
    service_name="analytics-service",
    module_id="analytics",
    service_version="1.0.0"
)

# Criar métricas customizadas
custom_metrics = create_custom_metrics()

app = FastAPI(title="Analytics Service API", version="1.0.0")

# Adicionar middleware para CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Adicionar middleware para contexto INNOVABIZ
app.add_middleware(InnovabizContextMiddleware)

# Modelo de dados para requisição
class AnalyticsRequest(BaseModel):
    event_type: str
    user_id: str
    data: Dict[str, Any]


# Dependência para obter contexto
async def get_innovabiz_context(request: Request):
    return request.scope.get("innovabiz_context", {
        "tenant_id": os.getenv("TENANT_ID", "default"),
        "region_id": os.getenv("REGION_ID", "default")
    })


@app.post("/api/v1/analytics/events")
async def record_event(
    event: AnalyticsRequest, 
    context: Dict[str, str] = Depends(get_innovabiz_context)
):
    # Registrar métricas de início
    start_time = time.time()
    
    # Incrementar contador de eventos com atributos multi-dimensionais
    custom_metrics["transaction_counter"].add(
        1, 
        {
            "innovabiz.tenant.id": context["tenant_id"],
            "innovabiz.region.id": context["region_id"],
            "event.type": event.event_type,
            "user.id": event.user_id[:8] if event.user_id else "anonymous"  # Truncar para privacidade
        }
    )
    
    # Lógica de processamento do evento...
    
    # Registrar duração do processamento
    duration = (time.time() - start_time) * 1000  # Converter para ms
    custom_metrics["transaction_duration"].record(
        duration,
        {
            "innovabiz.tenant.id": context["tenant_id"],
            "innovabiz.region.id": context["region_id"],
            "event.type": event.event_type
        }
    )
    
    return {
        "status": "success",
        "event_id": "evt-123456",  # Na implementação real, seria um ID gerado
        "tenant_id": context["tenant_id"],
        "region_id": context["region_id"]
    }


@app.get("/health")
async def health_check():
    """Endpoint de health check."""
    return {"status": "ok"}


@app.get("/metrics")
async def metrics():
    """Endpoint de métricas para compatibilidade."""
    return {"status": "metrics available via OpenTelemetry collector"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)
```

## Configuração de Variáveis de Ambiente

Crie um arquivo `.env` com as seguintes variáveis:

```
# Ambiente
ENVIRONMENT=development

# Contexto Multi-dimensional INNOVABIZ
TENANT_ID=default
REGION_ID=br

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://otel-collector:4317
OTEL_LOG_LEVEL=info
OTEL_RESOURCE_ATTRIBUTES=service.name=analytics-service,service.version=1.0.0,innovabiz.module.id=analytics

# Configurações de amostragem
OTEL_TRACES_SAMPLER=parentbased_traceidratio
OTEL_TRACES_SAMPLER_ARG=1.0

# Configurações de segurança
OTEL_EXPORTER_OTLP_HEADERS=x-innovabiz-tenant-id=default,x-innovabiz-region-id=br
```

## Melhores Práticas

1. **Nomenclatura de métricas**
   - Use `snake_case` para nomes de métricas
   - Siga o padrão `domínio.entidade.ação` (ex: `transactions.count`, `api.requests.duration`)
   - Use unidades padronizadas (ms, bytes, %, 1)

2. **Atributos obrigatórios para contexto multi-dimensional**
   - `innovabiz.tenant.id` - Identificador do tenant
   - `innovabiz.region.id` - Identificador da região
   - `innovabiz.module.id` - Identificador do módulo
   - `innovabiz.deployment.environment` - Ambiente de implantação

3. **Métricas essenciais por serviço**
   - Latência/duração das operações (histograms)
   - Contadores de operações (counters)
   - Taxa de erros (counters com atributos de status)
   - Utilização de recursos (gauges)
   - Estado do serviço (gauges)

4. **Propagação de contexto**
   - Propague sempre os cabeçalhos de contexto multi-dimensional entre serviços
   - Use middleware ASGI para consistência na gestão de contexto
   - Verifique sempre os cabeçalhos de entrada para extração do contexto

5. **Segurança e Compliance**
   - Não inclua dados sensíveis (PCI DSS, GDPR, LGPD) em métricas, logs ou traces
   - Utilize mascaramento ou truncamento de dados sensíveis (exemplo: truncar IDs de usuário)
   - Implemente controles de acesso (RBAC/ABAC) para visualização dos dados

## Checklist de Validação

- [ ] SDK OpenTelemetry inicializado antes de qualquer outro código
- [ ] Atributos de contexto multi-dimensional configurados corretamente
- [ ] Instrumentação automática configurada para todas as bibliotecas relevantes
- [ ] Métricas customizadas registradas conforme padrões INNOVABIZ
- [ ] Propagadores de contexto configurados corretamente
- [ ] Middleware para gestão de contexto multi-dimensional implementado
- [ ] Variáveis de ambiente documentadas
- [ ] Endpoints de health check implementados
- [ ] Dados sensíveis protegidos em conformidade com políticas de segurança
- [ ] Testes de verificação de telemetria implementados

## Recursos Adicionais

- [Documentação OpenTelemetry Python](https://opentelemetry.io/docs/instrumentation/python/)
- [Portal de Observabilidade INNOVABIZ](https://observability.innovabiz.com)
- [Repositório de Dashboards Padrão](https://github.com/innovabiz/observability-dashboards)
- [Guia de Troubleshooting](https://wiki.innovabiz.com/observability/troubleshooting)