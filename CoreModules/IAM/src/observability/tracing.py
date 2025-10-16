"""
Módulo de rastreamento distribuído com OpenTelemetry para o IAM Audit Service.

Implementa a integração com OpenTelemetry para fornecer rastreamento
distribuído, com suporte a contextos multi-tenant e multi-regionais.

Design baseado no ADR-008: Integração de Rastreamento Distribuído com OpenTelemetry.
"""

import functools
import inspect
import logging
import os
from typing import Dict, List, Optional, Callable, Any, Union, cast

from fastapi import FastAPI, Request, Response
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXInstrumentor
from opentelemetry.propagate import set_global_textmap, extract
from opentelemetry.propagators.b3 import B3MultiFormat
from opentelemetry.trace import Status, StatusCode, Span, set_span_in_context
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

from .config import TracingConfig

logger = logging.getLogger(__name__)


class TracingIntegration:
    """
    Integração de rastreamento distribuído com OpenTelemetry para o IAM Audit Service.
    
    Fornece instrumentação automática para FastAPI, clientes HTTP,
    e decoradores para rastreamento de funções de negócio.
    """
    
    def __init__(self, config: Optional[TracingConfig] = None):
        """
        Inicializa a integração de rastreamento.
        
        Args:
            config: Configuração para rastreamento
        """
        self.config = config or TracingConfig()
        self.tracer_provider = None
        self.tracer = None
        
        # Inicializa apenas se o rastreamento estiver habilitado
        if self.config.enabled:
            self._setup_tracing()
    
    def _setup_tracing(self):
        """Configura rastreamento com OpenTelemetry."""
        try:
            # Configura propagadores de contexto
            set_global_textmap(B3MultiFormat())
            
            # Cria recurso com atributos do serviço
            resource = Resource.create({
                "service.name": self.config.service_name,
                "service.namespace": self.config.namespace,
            })
            
            # Configura tracer provider
            self.tracer_provider = TracerProvider(resource=resource)
            
            # Configura exportador OTLP
            otlp_exporter = OTLPSpanExporter(
                endpoint=self.config.otlp_endpoint,
                headers=self.config.otlp_headers
            )
            
            # Adiciona processador de spans
            self.tracer_provider.add_span_processor(
                BatchSpanProcessor(otlp_exporter)
            )
            
            # Define tracer provider global
            trace.set_tracer_provider(self.tracer_provider)
            
            # Obtém tracer para o serviço
            self.tracer = trace.get_tracer(
                self.config.service_name,
                schema_url="https://opentelemetry.io/schemas/1.17.0"
            )
            
            # Inicializa instrumentações para clientes HTTP
            HTTPXInstrumentor().instrument()
            
            logger.info(
                f"OpenTelemetry tracing inicializado: "
                f"endpoint={self.config.otlp_endpoint}"
            )
            
        except Exception as e:
            logger.error(f"Falha ao configurar OpenTelemetry: {e}")
            # Desabilita tracing para evitar erros
            self.config.enabled = False
    
    def instrument_app(self, app: FastAPI):
        """
        Instrumenta uma aplicação FastAPI com rastreamento.
        
        Args:
            app: Aplicação FastAPI a ser instrumentada
        """
        if not self.config.enabled or not self.tracer_provider:
            logger.warning("Rastreamento não está habilitado. Ignorando instrumentação da app.")
            return
        
        try:
            # Instrumenta FastAPI
            FastAPIInstrumentor.instrument_app(
                app,
                tracer_provider=self.tracer_provider,
                excluded_urls=self.config.excluded_urls
            )
            
            # Adiciona middlewares para propagação de contexto
            # Isso é feito automaticamente pelo FastAPIInstrumentor
            
            # Adiciona handlers para spans customizados
            @app.middleware("http")
            async def add_tenant_region_attributes(request: Request, call_next):
                """
                Middleware para adicionar atributos de tenant e região ao span.
                
                Args:
                    request: Requisição HTTP
                    call_next: Próximo handler
                """
                # Extrai span atual
                current_span = trace.get_current_span()
                
                if current_span:
                    # Extrai contexto de headers
                    tenant = request.headers.get("X-Tenant-ID", "default")
                    region = request.headers.get("X-Region", "global")
                    environment = request.headers.get("X-Environment", "production")
                    
                    # Adiciona atributos ao span
                    current_span.set_attribute("tenant.id", tenant)
                    current_span.set_attribute("region.code", region)
                    current_span.set_attribute("deployment.environment", environment)
                
                # Continua processamento
                response = await call_next(request)
                
                # Adiciona informações de resposta se disponíveis
                if current_span and response:
                    current_span.set_attribute("http.status_code", response.status_code)
                
                return response
            
            logger.info("FastAPI instrumentada com OpenTelemetry")
            
        except Exception as e:
            logger.error(f"Falha ao instrumentar FastAPI com OpenTelemetry: {e}")
    
    async def initialize(self):
        """Inicializa recursos de rastreamento."""
        # Nada a fazer por enquanto
        pass
    
    async def shutdown(self):
        """Limpa recursos de rastreamento."""
        if self.tracer_provider:
            self.tracer_provider.shutdown()
    
    def create_span(
        self,
        name: str,
        context: Optional[Any] = None,
        kind: Optional[trace.SpanKind] = trace.SpanKind.INTERNAL,
        attributes: Optional[Dict[str, Any]] = None
    ) -> Span:
        """
        Cria um novo span.
        
        Args:
            name: Nome do span
            context: Contexto do span (opcional)
            kind: Tipo do span
            attributes: Atributos adicionais para o span
            
        Returns:
            Novo span
        """
        if not self.config.enabled or not self.tracer:
            # Retorna um span noop se o rastreamento estiver desabilitado
            return trace.INVALID_SPAN
        
        return self.tracer.start_as_current_span(
            name,
            context=context,
            kind=kind,
            attributes=attributes or {}
        )
    
    def extract_context(self, headers: Dict[str, str]):
        """
        Extrai contexto de rastreamento de headers HTTP.
        
        Args:
            headers: Headers HTTP
            
        Returns:
            Contexto de rastreamento
        """
        if not self.config.enabled:
            return None
        
        return extract(headers)
    
    def inject_context(self, headers: Dict[str, str]):
        """
        Injeta contexto de rastreamento em headers HTTP.
        
        Args:
            headers: Headers HTTP a serem atualizados
        """
        if not self.config.enabled:
            return
        
        # Obtém propagador B3
        propagator = B3MultiFormat()
        
        # Injeta contexto nos headers
        propagator.inject(headers)


def traced(
    name: Optional[str] = None,
    kind: Optional[trace.SpanKind] = trace.SpanKind.INTERNAL,
    attributes: Optional[Dict[str, Any]] = None,
    record_exception: bool = True
):
    """
    Decorador para rastrear funções e métodos.
    
    Args:
        name: Nome do span (opcional, usa nome da função se não for fornecido)
        kind: Tipo do span
        attributes: Atributos estáticos para adicionar ao span
        record_exception: Se deve registrar exceções no span
        
    Returns:
        Decorador configurado
    """
    def decorator(func):
        # Mantém assinatura e docstrings
        @functools.wraps(func)
        async def async_wrapper(*args, **kwargs):
            # Obtém o nome do span (nome da função se não fornecido)
            span_name = name or f"{func.__module__}.{func.__qualname__}"
            
            # Prepara atributos
            span_attrs = attributes.copy() if attributes else {}
            
            # Extrai contexto de tenant e região, se disponíveis
            if args and hasattr(args[0], "tenant") and hasattr(args[0], "region"):
                span_attrs["tenant.id"] = args[0].tenant
                span_attrs["region.code"] = args[0].region
            
            # Para métodos com request como argumento
            for arg in args:
                if hasattr(arg, "headers"):
                    # Extrai de headers em um objeto request
                    headers = getattr(arg, "headers")
                    tenant = headers.get("X-Tenant-ID", "default")
                    region = headers.get("X-Region", "global")
                    span_attrs["tenant.id"] = tenant
                    span_attrs["region.code"] = region
                    break
            
            # Extrai de kwargs
            if "tenant" in kwargs:
                span_attrs["tenant.id"] = kwargs["tenant"]
            
            if "region" in kwargs:
                span_attrs["region.code"] = kwargs["region"]
            
            # Cria span
            tracer = trace.get_tracer(__name__)
            with tracer.start_as_current_span(
                span_name,
                kind=kind,
                attributes=span_attrs
            ) as span:
                try:
                    result = await func(*args, **kwargs)
                    return result
                except Exception as e:
                    if record_exception:
                        span.record_exception(e)
                        span.set_status(Status(StatusCode.ERROR, str(e)))
                    raise
        
        @functools.wraps(func)
        def sync_wrapper(*args, **kwargs):
            # Obtém o nome do span (nome da função se não fornecido)
            span_name = name or f"{func.__module__}.{func.__qualname__}"
            
            # Prepara atributos
            span_attrs = attributes.copy() if attributes else {}
            
            # Extrai contexto de tenant e região, se disponíveis
            if args and hasattr(args[0], "tenant") and hasattr(args[0], "region"):
                span_attrs["tenant.id"] = args[0].tenant
                span_attrs["region.code"] = args[0].region
            
            # Para métodos com request como argumento
            for arg in args:
                if hasattr(arg, "headers"):
                    # Extrai de headers em um objeto request
                    headers = getattr(arg, "headers")
                    tenant = headers.get("X-Tenant-ID", "default")
                    region = headers.get("X-Region", "global")
                    span_attrs["tenant.id"] = tenant
                    span_attrs["region.code"] = region
                    break
            
            # Extrai de kwargs
            if "tenant" in kwargs:
                span_attrs["tenant.id"] = kwargs["tenant"]
            
            if "region" in kwargs:
                span_attrs["region.code"] = kwargs["region"]
            
            # Cria span
            tracer = trace.get_tracer(__name__)
            with tracer.start_as_current_span(
                span_name,
                kind=kind,
                attributes=span_attrs
            ) as span:
                try:
                    result = func(*args, **kwargs)
                    return result
                except Exception as e:
                    if record_exception:
                        span.record_exception(e)
                        span.set_status(Status(StatusCode.ERROR, str(e)))
                    raise
        
        # Retorna wrapper apropriado (assíncrono ou síncrono)
        if inspect.iscoroutinefunction(func):
            return async_wrapper
        return sync_wrapper
    
    return decorator


def trace_audit_event(event_type: str):
    """
    Decorador específico para rastrear eventos de auditoria.
    
    Args:
        event_type: Tipo do evento de auditoria
        
    Returns:
        Decorador configurado
    """
    # Configura amostragem específica para o tipo de evento
    sample_ratio = 1.0  # Valor padrão
    
    # Tenta obter configuração do singleton TracingIntegration
    # Este é um exemplo - implementação completa depende da estrutura global
    config = TracingConfig()  # Fallback para configuração padrão
    
    # Verifica se há configuração específica para o tipo de evento
    if hasattr(config, "span_configs") and event_type in config.span_configs:
        event_config = config.span_configs[event_type]
        sample_ratio = event_config.get("sample_ratio", 1.0)
    
    # Determina se deve amostrar este evento específico
    # Em uma implementação real, usaria um sampler adequado
    import random
    should_sample = random.random() < sample_ratio
    
    if not should_sample:
        # Se não for amostrar, retorna um decorador transparente
        def transparent_decorator(func):
            return func
        return transparent_decorator
    
    # Se for amostrar, utiliza o decorador traced
    return traced(
        name=f"audit.event.{event_type}",
        kind=trace.SpanKind.CONSUMER,
        attributes={
            "audit.event_type": event_type,
            "messaging.system": "audit"
        }
    )


def trace_compliance_check(compliance_type: str):
    """
    Decorador específico para rastrear verificações de compliance.
    
    Args:
        compliance_type: Tipo da verificação de compliance
        
    Returns:
        Decorador configurado
    """
    return traced(
        name=f"compliance.check.{compliance_type}",
        kind=trace.SpanKind.INTERNAL,
        attributes={
            "compliance.type": compliance_type
        }
    )


def trace_retention_policy(policy_name: str):
    """
    Decorador específico para rastrear execução de políticas de retenção.
    
    Args:
        policy_name: Nome da política de retenção
        
    Returns:
        Decorador configurado
    """
    return traced(
        name=f"retention.execution.{policy_name}",
        kind=trace.SpanKind.INTERNAL,
        attributes={
            "retention.policy_name": policy_name
        }
    )


# Exemplo de uso dos decoradores
"""
class AuditEventProcessor:
    
    @trace_audit_event("login")
    async def process_login_event(self, event_data: Dict[str, Any], tenant: str, region: str):
        # Processamento do evento de login
        pass
    
    @trace_compliance_check("pci_dss")
    async def check_pci_compliance(self, audit_record: AuditRecord, tenant: str):
        # Verifica conformidade com PCI DSS
        pass
    
    @trace_retention_policy("gdpr_90_days")
    async def apply_gdpr_retention(self, tenant: str, region: str):
        # Aplicação de política de retenção GDPR (90 dias)
        pass
"""