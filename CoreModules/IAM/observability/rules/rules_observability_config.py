"""
Configuração de observabilidade para o sistema de regras dinâmicas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/Observabilidade
Data: 21/08/2025
"""

import logging
import os
from typing import Dict, Optional, List, Any

from opentelemetry import trace, metrics
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metrics_exporter import OTLPMetricsExporter
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, DEPLOYMENT_ENVIRONMENT
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator

from pydantic import BaseModel, Field


class ObservabilityConfig(BaseModel):
    """Configuração para observabilidade."""

    # Configuração geral
    service_name: str = Field("innovabiz-iam-rules", description="Nome do serviço")
    environment: str = Field("development", description="Ambiente (development, staging, production)")
    version: str = Field("1.0.0", description="Versão do serviço")
    tenant_id: str = Field("default", description="ID do tenant")
    
    # Configuração de logging
    log_level: str = Field("INFO", description="Nível de log (DEBUG, INFO, WARNING, ERROR, CRITICAL)")
    log_format: str = Field(
        "%(asctime)s - %(name)s - %(levelname)s - [%(tenant_id)s] - %(message)s",
        description="Formato do log"
    )
    log_file: Optional[str] = Field(None, description="Arquivo de log (opcional)")
    
    # Configuração de tracing
    tracing_enabled: bool = Field(True, description="Habilitar tracing")
    tracing_endpoint: str = Field(
        "http://otel-collector:4317", 
        description="Endpoint do coletor OpenTelemetry para tracing"
    )
    tracing_sample_rate: float = Field(1.0, description="Taxa de amostragem para tracing (0.0 - 1.0)")
    
    # Configuração de métricas
    metrics_enabled: bool = Field(True, description="Habilitar métricas")
    metrics_endpoint: str = Field(
        "http://otel-collector:4317",
        description="Endpoint do coletor OpenTelemetry para métricas"
    )
    metrics_export_interval_ms: int = Field(
        60000, 
        description="Intervalo de exportação de métricas em milissegundos"
    )
    
    # Metadados adicionais
    tags: Dict[str, str] = Field(
        default_factory=dict,
        description="Tags adicionais para telemetria"
    )
    
    class Config:
        """Configuração do Pydantic."""
        validate_assignment = True
        extra = "allow"


class RulesObservabilityConfigurator:
    """
    Configurador de observabilidade para o sistema de regras dinâmicas.
    
    Provê configuração integrada para:
    1. Logging
    2. Tracing (OpenTelemetry)
    3. Métricas (OpenTelemetry)
    """
    
    def __init__(
        self,
        config: ObservabilityConfig,
        logger_name: str = "innovabiz.iam.rules",
    ):
        """
        Inicializa o configurador.
        
        Args:
            config: Configuração de observabilidade
            logger_name: Nome do logger
        """
        self.config = config
        self.logger_name = logger_name
        self._logger = None
        self._tracer = None
        self._meter = None
        
        # Inicializar recursos
        self._setup_resources()
    
    def _setup_resources(self):
        """Configura recursos compartilhados por tracer e meter."""
        self.resource = Resource.create({
            SERVICE_NAME: self.config.service_name,
            DEPLOYMENT_ENVIRONMENT: self.config.environment,
            "service.version": self.config.version,
            "tenant.id": self.config.tenant_id,
            **self.config.tags,
        })
    
    def setup_logger(self) -> logging.Logger:
        """
        Configura e retorna um logger.
        
        Returns:
            logging.Logger: Logger configurado
        """
        if self._logger is not None:
            return self._logger
        
        # Obter o logger
        logger = logging.getLogger(self.logger_name)
        
        # Configurar nível de log
        level = getattr(logging, self.config.log_level.upper(), logging.INFO)
        logger.setLevel(level)
        
        # Criar handler para console
        console_handler = logging.StreamHandler()
        console_handler.setLevel(level)
        
        # Criar formatador
        log_format = self.config.log_format
        formatter = logging.Formatter(log_format)
        console_handler.setFormatter(formatter)
        
        # Adicionar handler ao logger
        logger.addHandler(console_handler)
        
        # Adicionar handler de arquivo, se configurado
        if self.config.log_file:
            file_handler = logging.FileHandler(self.config.log_file)
            file_handler.setLevel(level)
            file_handler.setFormatter(formatter)
            logger.addHandler(file_handler)
        
        # Filtro para adicionar tenant_id aos logs
        class TenantFilter(logging.Filter):
            def __init__(self, tenant_id):
                super().__init__()
                self.tenant_id = tenant_id
            
            def filter(self, record):
                record.tenant_id = self.tenant_id
                return True
        
        tenant_filter = TenantFilter(self.config.tenant_id)
        logger.addFilter(tenant_filter)
        
        # Armazenar e retornar logger
        self._logger = logger
        return logger
    
    def setup_tracer(self) -> trace.Tracer:
        """
        Configura e retorna um tracer OpenTelemetry.
        
        Returns:
            trace.Tracer: Tracer configurado
        """
        if not self.config.tracing_enabled:
            # Retornar um tracer noop
            return trace.get_tracer(self.config.service_name)
        
        if self._tracer is not None:
            return self._tracer
        
        # Configurar provedor de tracing
        trace_provider = TracerProvider(
            resource=self.resource,
        )
        
        # Configurar exportador
        otlp_exporter = OTLPSpanExporter(endpoint=self.config.tracing_endpoint)
        span_processor = BatchSpanProcessor(otlp_exporter)
        trace_provider.add_span_processor(span_processor)
        
        # Registrar o provedor
        trace.set_tracer_provider(trace_provider)
        
        # Criar e retornar tracer
        self._tracer = trace.get_tracer(
            self.config.service_name,
            self.config.version,
        )
        return self._tracer
    
    def setup_meter(self) -> metrics.Meter:
        """
        Configura e retorna um meter OpenTelemetry.
        
        Returns:
            metrics.Meter: Meter configurado
        """
        if not self.config.metrics_enabled:
            # Retornar um meter noop
            return metrics.get_meter(self.config.service_name)
        
        if self._meter is not None:
            return self._meter
        
        # Configurar exportador de métricas
        otlp_exporter = OTLPMetricsExporter(endpoint=self.config.metrics_endpoint)
        reader = PeriodicExportingMetricReader(
            exporter=otlp_exporter,
            export_interval_millis=self.config.metrics_export_interval_ms,
        )
        
        # Configurar provedor de métricas
        metrics_provider = MeterProvider(
            resource=self.resource,
            metric_readers=[reader],
        )
        
        # Registrar o provedor
        metrics.set_meter_provider(metrics_provider)
        
        # Criar e retornar meter
        self._meter = metrics.get_meter(
            self.config.service_name,
            self.config.version,
        )
        return self._meter
    
    def get_trace_context_carrier(self, span) -> Dict[str, str]:
        """
        Retorna um carrier para propagação de contexto de trace.
        
        Args:
            span: Span atual
            
        Returns:
            Dict[str, str]: Carrier com contexto de trace
        """
        carrier = {}
        propagator = TraceContextTextMapPropagator()
        propagator.inject(carrier=carrier)
        return carrier
    
    @classmethod
    def from_env(cls, logger_name: str = "innovabiz.iam.rules") -> 'RulesObservabilityConfigurator':
        """
        Cria um configurador a partir de variáveis de ambiente.
        
        Args:
            logger_name: Nome do logger
            
        Returns:
            RulesObservabilityConfigurator: Configurador inicializado
        """
        config = ObservabilityConfig(
            service_name=os.environ.get("OTEL_SERVICE_NAME", "innovabiz-iam-rules"),
            environment=os.environ.get("OTEL_ENVIRONMENT", "development"),
            version=os.environ.get("OTEL_VERSION", "1.0.0"),
            tenant_id=os.environ.get("TENANT_ID", "default"),
            log_level=os.environ.get("LOG_LEVEL", "INFO"),
            log_format=os.environ.get("LOG_FORMAT", 
                                      "%(asctime)s - %(name)s - %(levelname)s - "
                                      "[%(tenant_id)s] - %(message)s"),
            log_file=os.environ.get("LOG_FILE"),
            tracing_enabled=os.environ.get("TRACING_ENABLED", "true").lower() == "true",
            tracing_endpoint=os.environ.get("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", 
                                           "http://otel-collector:4317"),
            tracing_sample_rate=float(os.environ.get("TRACING_SAMPLE_RATE", "1.0")),
            metrics_enabled=os.environ.get("METRICS_ENABLED", "true").lower() == "true",
            metrics_endpoint=os.environ.get("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", 
                                          "http://otel-collector:4317"),
            metrics_export_interval_ms=int(os.environ.get("METRICS_EXPORT_INTERVAL_MS", "60000")),
        )
        
        # Adicionar tags de ambiente
        for key, value in os.environ.items():
            if key.startswith("OTEL_TAG_"):
                tag_name = key.replace("OTEL_TAG_", "").lower()
                config.tags[tag_name] = value
        
        return cls(config, logger_name)