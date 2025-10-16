"""
INNOVABIZ IAM - Observability Module
Autor: Eduardo Jeremias
Versão: 1.0.0
Descrição: Módulo central de observabilidade para o sistema IAM
Compatibilidade: Multi-contexto (BR, US, EU, AO)
Compliance: GDPR, LGPD, PCI DSS 4.0, PSD2, BACEN, BNA

Este módulo implementa as melhores práticas para observabilidade:
1. Tracing distribuído (OpenTelemetry)
2. Métricas (Prometheus)
3. Logging estruturado (JSON)
4. Correlação entre componentes
"""

import os
import json
import logging
import time
from typing import Dict, Any, Optional
from functools import lru_cache
from contextlib import contextmanager
import structlog
from prometheus_client import Counter, Histogram, Gauge, Summary, CollectorRegistry
import opentelemetry.trace as trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.trace import SpanKind, Status, StatusCode

# Configuração do Logger Estruturado
structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.JSONRenderer()
    ],
    logger_factory=structlog.stdlib.LoggerFactory(),
)

# Configurar logger padrão para usar o formato estruturado
logging.basicConfig(
    format="%(message)s",
    level=logging.INFO,
)

# Logger estruturado
logger = structlog.get_logger("innovabiz.iam")

# Configuração de Tracing
class TracingSetup:
    """Configuração do sistema de tracing distribuído com OpenTelemetry"""
    
    def __init__(self):
        """Inicializa o sistema de tracing"""
        self.service_name = os.environ.get("SERVICE_NAME", "innovabiz-iam")
        self.service_version = os.environ.get("SERVICE_VERSION", "1.0.0")
        self.environment = os.environ.get("ENVIRONMENT", "development")
        
        # Definir recurso para identificação do serviço
        resource = Resource.create({
            "service.name": self.service_name,
            "service.version": self.service_version,
            "environment": self.environment
        })
        
        # Configurar provedor de trace
        provider = TracerProvider(resource=resource)
        
        # Configurar exportador OTLP se disponível
        otlp_endpoint = os.environ.get("OTLP_ENDPOINT")
        if otlp_endpoint:
            otlp_exporter = OTLPSpanExporter(endpoint=otlp_endpoint)
            span_processor = BatchSpanProcessor(otlp_exporter)
            provider.add_span_processor(span_processor)
        
        # Registrar provedor global
        trace.set_tracer_provider(provider)
        
        # Criar tracer
        self.tracer = trace.get_tracer(
            self.service_name,
            self.service_version
        )
    
    @contextmanager
    def start_span(self, name: str, context: Dict[str, Any] = None, kind: SpanKind = SpanKind.INTERNAL):
        """
        Inicia um span de tracing.
        
        Args:
            name: Nome do span
            context: Atributos de contexto
            kind: Tipo de span
            
        Returns:
            Span context manager
        """
        context = context or {}
        with self.tracer.start_as_current_span(name, kind=kind, attributes=context) as span:
            try:
                yield span
                span.set_status(Status(StatusCode.OK))
            except Exception as e:
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                raise
    
    def add_span_event(self, span, name: str, attributes: Dict[str, Any] = None):
        """
        Adiciona um evento a um span existente.
        
        Args:
            span: Span atual
            name: Nome do evento
            attributes: Atributos do evento
        """
        attributes = attributes or {}
        span.add_event(name, attributes=attributes)

# Inicialização do sistema de tracing
tracing = TracingSetup()

# Configuração das métricas Prometheus
class PrometheusMetrics:
    """Métricas Prometheus para o sistema IAM"""
    
    def __init__(self):
        """Inicializa as métricas Prometheus"""
        self.registry = CollectorRegistry()
        
        # Métricas de auditoria
        self.audit_events_total = Counter(
            "innovabiz_iam_audit_events_total",
            "Total de eventos de auditoria por categoria, severidade, tenant e contexto regional",
            ["category", "severity", "tenant_id", "regional_context"],
            registry=self.registry
        )
        
        self.audit_api_requests = Counter(
            "innovabiz_iam_audit_api_requests_total",
            "Total de requisições à API de auditoria",
            ["endpoint", "tenant_id", "regional_context"],
            registry=self.registry
        )
        
        self.audit_processing_time = Histogram(
            "innovabiz_iam_audit_processing_time_seconds",
            "Tempo de processamento de eventos de auditoria",
            ["operation", "tenant_id", "regional_context"],
            registry=self.registry,
            buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0)
        )
        
        self.audit_exports = Counter(
            "innovabiz_iam_audit_exports_total",
            "Total de exportações de dados de auditoria por formato",
            ["format", "tenant_id"],
            registry=self.registry
        )
        
        self.audit_storage_size = Gauge(
            "innovabiz_iam_audit_storage_size_bytes",
            "Tamanho do armazenamento de auditoria em bytes",
            ["tenant_id"],
            registry=self.registry
        )
        
        # Métricas HTTP
        self.http_requests_total = Counter(
            "innovabiz_iam_http_requests_total",
            "Total de requisições HTTP",
            ["method", "path", "status_code", "tenant_id", "regional_context"],
            registry=self.registry
        )
        
        self.http_request_duration = Histogram(
            "innovabiz_iam_http_request_duration_seconds",
            "Duração das requisições HTTP em segundos",
            ["method", "path", "status_code", "tenant_id", "regional_context"],
            registry=self.registry,
            buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0)
        )
        
        self.http_request_errors_total = Counter(
            "innovabiz_iam_http_request_errors_total",
            "Total de erros em requisições HTTP",
            ["method", "path", "exception_type", "tenant_id", "regional_context"],
            registry=self.registry
        )
        
        # Métricas de banco de dados
        self.db_operation_duration = Histogram(
            "innovabiz_iam_db_operation_duration_seconds",
            "Duração de operações de banco de dados em segundos",
            ["operation", "collection", "tenant_id"],
            registry=self.registry,
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5)
        )
        
        self.db_connections = Gauge(
            "innovabiz_iam_db_connections",
            "Número de conexões ativas com o banco de dados",
            ["db_type"],
            registry=self.registry
        )
        
        # Métricas Redis
        self.redis_operation_duration = Histogram(
            "innovabiz_iam_redis_operation_duration_seconds",
            "Duração de operações Redis em segundos",
            ["operation"],
            registry=self.registry,
            buckets=(0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25)
        )
        
        self.redis_connections = Gauge(
            "innovabiz_iam_redis_connections",
            "Número de conexões ativas com o Redis",
            [],
            registry=self.registry
        )
    
    @contextmanager
    def timed_execution(self, metric, labels=None):
        """
        Mede o tempo de execução de um bloco de código.
        
        Args:
            metric: Métrica Histogram a ser usada
            labels: Labels para a métrica
            
        Returns:
            Context manager para medição de tempo
        """
        labels = labels or {}
        start_time = time.time()
        try:
            yield
        finally:
            end_time = time.time()
            duration = end_time - start_time
            metric.labels(**labels).observe(duration)

# Inicialização das métricas Prometheus
metrics = PrometheusMetrics()

class ObservabilityContext:
    """
    Contexto de observabilidade para o sistema IAM.
    
    Responsável por:
    - Correlacionar logs, métricas e traces
    - Enriquecer logs com contexto
    - Medir tempos de operações
    """
    
    def __init__(self, 
                 trace_id: Optional[str] = None,
                 span_id: Optional[str] = None,
                 regional_context: Optional[str] = None,
                 tenant_id: Optional[str] = None,
                 user_id: Optional[str] = None,
                 request_id: Optional[str] = None):
        """
        Inicializa um contexto de observabilidade.
        
        Args:
            trace_id: ID do trace distribuído
            span_id: ID do span atual
            regional_context: Contexto regional (BR, US, EU, AO)
            tenant_id: ID do tenant
            user_id: ID do usuário
            request_id: ID da requisição
        """
        self.trace_id = trace_id
        self.span_id = span_id
        self.regional_context = regional_context
        self.tenant_id = tenant_id
        self.user_id = user_id
        self.request_id = request_id
        
        # Construir contexto para logs estruturados
        self.log_context = {}
        if trace_id:
            self.log_context["trace_id"] = trace_id
        if span_id:
            self.log_context["span_id"] = span_id
        if regional_context:
            self.log_context["regional_context"] = regional_context
        if tenant_id:
            self.log_context["tenant_id"] = tenant_id
        if user_id:
            self.log_context["user_id"] = user_id
        if request_id:
            self.log_context["request_id"] = request_id
    
    def bind_logger(self):
        """
        Cria um logger vinculado ao contexto atual.
        
        Returns:
            Logger vinculado ao contexto
        """
        return logger.bind(**self.log_context)
    
    def start_span(self, name: str, extra_context: Dict[str, Any] = None):
        """
        Inicia um span no contexto atual.
        
        Args:
            name: Nome do span
            extra_context: Contexto adicional
            
        Returns:
            Span context manager
        """
        context = self.log_context.copy()
        if extra_context:
            context.update(extra_context)
        
        return tracing.start_span(name, context)
    
    def timed_operation(self, metric, labels=None):
        """
        Mede o tempo de uma operação.
        
        Args:
            metric: Métrica a ser atualizada
            labels: Labels para a métrica
            
        Returns:
            Context manager para medição de tempo
        """
        labels_dict = labels or {}
        
        # Adicionar contexto às labels
        if self.tenant_id and "tenant_id" not in labels_dict:
            labels_dict["tenant_id"] = self.tenant_id
            
        if self.regional_context and "regional_context" not in labels_dict:
            labels_dict["regional_context"] = self.regional_context
            
        return metrics.timed_execution(metric, labels_dict)

@lru_cache(maxsize=1024)
def get_observability_context(
    trace_id: Optional[str] = None,
    span_id: Optional[str] = None,
    regional_context: Optional[str] = None,
    tenant_id: Optional[str] = None,
    user_id: Optional[str] = None,
    request_id: Optional[str] = None
) -> ObservabilityContext:
    """
    Obtém um contexto de observabilidade (cacheado para performance).
    
    Args:
        trace_id: ID do trace distribuído
        span_id: ID do span atual
        regional_context: Contexto regional (BR, US, EU, AO)
        tenant_id: ID do tenant
        user_id: ID do usuário
        request_id: ID da requisição
        
    Returns:
        Contexto de observabilidade
    """
    return ObservabilityContext(
        trace_id=trace_id,
        span_id=span_id,
        regional_context=regional_context,
        tenant_id=tenant_id,
        user_id=user_id,
        request_id=request_id
    )