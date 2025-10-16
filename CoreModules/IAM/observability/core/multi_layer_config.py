"""
Configuração de Observabilidade Multi-Camada para a plataforma INNOVABIZ.

Este módulo implementa a configuração centralizada para observabilidade
multi-camada, multi-tenant e multi-contexto, fornecendo suporte para
telemetria, métricas, logs e traces para todos os módulos da plataforma.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ - Sistema de Observabilidade Multi-Camada
Data: 20/08/2025
"""

import os
import json
import logging
from typing import Dict, List, Optional, Union, Any
from enum import Enum
from datetime import datetime

from pydantic import BaseModel, Field, validator

# Configurações para os diferentes níveis de observabilidade
class ObservabilityLevel(str, Enum):
    """Níveis de observabilidade para componentes do sistema."""
    BASIC = "basic"           # Logs básicos e métricas essenciais
    STANDARD = "standard"     # Logs detalhados, métricas e traces básicos
    ADVANCED = "advanced"     # Telemetria completa, logs detalhados, traces e métricas avançadas
    DIAGNOSTIC = "diagnostic" # Modo diagnóstico completo (alta verbosidade, rastreamento completo)


class SecurityContext(BaseModel):
    """Contexto de segurança para telemetria."""
    tenant_id: str
    module_id: str
    component_id: str
    confidentiality_level: str = "standard"
    data_classification: str = "internal"
    compliance_contexts: List[str] = Field(default_factory=list)
    regulatory_jurisdictions: List[str] = Field(default_factory=list)


class TelemetryTag(BaseModel):
    """Tag de telemetria para enriquecimento de dados."""
    key: str
    value: str
    context: Optional[str] = None


class MultiLayerObservabilityConfig(BaseModel):
    """
    Configuração centralizada para observabilidade multi-camada.
    
    Suporta configuração para diferentes camadas da plataforma:
    - Infraestrutura (servidores, redes, containers)
    - Aplicação (serviços, APIs)
    - Negócio (transações, processos de negócio)
    - Usuário (experiência do usuário, interações)
    - Segurança (eventos de segurança, ameaças)
    """
    # Metadados do serviço
    service_name: str
    environment: str = "development"
    version: str = "1.0.0"
    module_name: str
    component_name: Optional[str] = None
    
    # Contextos específicos
    tenant_id: str
    region: Optional[str] = None
    market_context: Optional[str] = None
    
    # Níveis de observabilidade por camada
    infrastructure_layer_level: ObservabilityLevel = ObservabilityLevel.STANDARD
    application_layer_level: ObservabilityLevel = ObservabilityLevel.STANDARD
    business_layer_level: ObservabilityLevel = ObservabilityLevel.STANDARD
    user_layer_level: ObservabilityLevel = ObservabilityLevel.STANDARD
    security_layer_level: ObservabilityLevel = ObservabilityLevel.ADVANCED
    
    # Configurações de logging
    log_level: str = "INFO"
    log_format: str = "json"
    log_file: Optional[str] = None
    structured_logging: bool = True
    tenant_filtered_logging: bool = True
    
    # Configurações de tracing
    tracing_enabled: bool = True
    tracing_sample_rate: float = 0.1
    otel_exporter_otlp_traces_endpoint: Optional[str] = None
    trace_propagation_format: str = "w3c"
    
    # Configurações de métricas
    metrics_enabled: bool = True
    metrics_export_interval_ms: int = 60000
    otel_exporter_otlp_metrics_endpoint: Optional[str] = None
    high_cardinality_metrics: bool = False
    
    # Configurações de segurança e compliance
    security_context: Optional[SecurityContext] = None
    pii_masking_enabled: bool = True
    compliance_tracking_enabled: bool = True
    regulatory_frameworks: List[str] = Field(default_factory=list)
    
    # Tags de telemetria adicionais
    telemetry_tags: List[TelemetryTag] = Field(default_factory=list)
    
    # Cache e otimizações
    telemetry_batching: bool = True
    batch_size: int = 100
    telemetry_compression: bool = True
    
    @validator('service_name')
    def validate_service_name(cls, v):
        """Valida o nome do serviço."""
        if not v or len(v.strip()) == 0:
            raise ValueError("O nome do serviço não pode ser vazio")
        return v
    
    @validator('tenant_id')
    def validate_tenant_id(cls, v):
        """Valida o ID do tenant."""
        if not v or len(v.strip()) == 0:
            raise ValueError("O ID do tenant não pode ser vazio")
        return v
    
    @classmethod
    def from_env(cls, logger_name: str = "innovabiz.observability") -> 'MultiLayerObservabilityConfig':
        """
        Cria uma configuração de observabilidade a partir de variáveis de ambiente.
        
        Args:
            logger_name: Nome do logger a ser criado
        
        Returns:
            MultiLayerObservabilityConfig: Configuração de observabilidade
        """
        # Configurações básicas
        config_dict = {
            "service_name": os.environ.get("OTEL_SERVICE_NAME", "innovabiz-service"),
            "environment": os.environ.get("OTEL_ENVIRONMENT", "development"),
            "version": os.environ.get("OTEL_VERSION", "1.0.0"),
            "module_name": os.environ.get("MODULE_NAME", "undefined"),
            "component_name": os.environ.get("COMPONENT_NAME", None),
            "tenant_id": os.environ.get("TENANT_ID", "default"),
            "region": os.environ.get("REGION", None),
            "market_context": os.environ.get("MARKET_CONTEXT", None),
        }
        
        # Níveis de observabilidade
        config_dict.update({
            "infrastructure_layer_level": os.environ.get("INFRASTRUCTURE_LAYER_LEVEL", ObservabilityLevel.STANDARD),
            "application_layer_level": os.environ.get("APPLICATION_LAYER_LEVEL", ObservabilityLevel.STANDARD),
            "business_layer_level": os.environ.get("BUSINESS_LAYER_LEVEL", ObservabilityLevel.STANDARD),
            "user_layer_level": os.environ.get("USER_LAYER_LEVEL", ObservabilityLevel.STANDARD),
            "security_layer_level": os.environ.get("SECURITY_LAYER_LEVEL", ObservabilityLevel.ADVANCED),
        })
        
        # Configurações de logging
        config_dict.update({
            "log_level": os.environ.get("LOG_LEVEL", "INFO"),
            "log_format": os.environ.get("LOG_FORMAT", "json"),
            "log_file": os.environ.get("LOG_FILE", None),
            "structured_logging": os.environ.get("STRUCTURED_LOGGING", "true").lower() == "true",
            "tenant_filtered_logging": os.environ.get("TENANT_FILTERED_LOGGING", "true").lower() == "true",
        })
        
        # Configurações de tracing
        config_dict.update({
            "tracing_enabled": os.environ.get("TRACING_ENABLED", "true").lower() == "true",
            "tracing_sample_rate": float(os.environ.get("TRACING_SAMPLE_RATE", "0.1")),
            "otel_exporter_otlp_traces_endpoint": os.environ.get("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", None),
            "trace_propagation_format": os.environ.get("TRACE_PROPAGATION_FORMAT", "w3c"),
        })
        
        # Configurações de métricas
        config_dict.update({
            "metrics_enabled": os.environ.get("METRICS_ENABLED", "true").lower() == "true",
            "metrics_export_interval_ms": int(os.environ.get("METRICS_EXPORT_INTERVAL_MS", "60000")),
            "otel_exporter_otlp_metrics_endpoint": os.environ.get("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", None),
            "high_cardinality_metrics": os.environ.get("HIGH_CARDINALITY_METRICS", "false").lower() == "true",
        })
        
        # Configurações de segurança e compliance
        regulatory_frameworks = []
        if os.environ.get("REGULATORY_FRAMEWORKS"):
            try:
                regulatory_frameworks = json.loads(os.environ.get("REGULATORY_FRAMEWORKS", "[]"))
            except json.JSONDecodeError:
                regulatory_frameworks = os.environ.get("REGULATORY_FRAMEWORKS", "").split(",")
        
        config_dict.update({
            "pii_masking_enabled": os.environ.get("PII_MASKING_ENABLED", "true").lower() == "true",
            "compliance_tracking_enabled": os.environ.get("COMPLIANCE_TRACKING_ENABLED", "true").lower() == "true",
            "regulatory_frameworks": regulatory_frameworks,
        })
        
        # Security context
        if os.environ.get("SECURITY_CONTEXT"):
            try:
                sec_context = json.loads(os.environ.get("SECURITY_CONTEXT", "{}"))
                config_dict["security_context"] = SecurityContext(**sec_context)
            except (json.JSONDecodeError, ValueError):
                pass
        
        # Telemetry tags
        telemetry_tags = []
        for key, value in os.environ.items():
            if key.startswith("OTEL_TAG_"):
                tag_key = key.replace("OTEL_TAG_", "").lower()
                telemetry_tags.append(TelemetryTag(key=tag_key, value=value))
        
        config_dict["telemetry_tags"] = telemetry_tags
        
        # Cache e otimizações
        config_dict.update({
            "telemetry_batching": os.environ.get("TELEMETRY_BATCHING", "true").lower() == "true",
            "batch_size": int(os.environ.get("TELEMETRY_BATCH_SIZE", "100")),
            "telemetry_compression": os.environ.get("TELEMETRY_COMPRESSION", "true").lower() == "true",
        })
        
        return cls(**config_dict)


class MultiLayerObservabilityConfigurator:
    """
    Configurador para observabilidade multi-camada.
    
    Responsável por inicializar e configurar componentes de observabilidade
    para diferentes camadas do sistema, aplicando as configurações adequadas
    para cada contexto (tenant, região, ambiente, etc.).
    """
    
    def __init__(self, config: MultiLayerObservabilityConfig, logger_name: str):
        """
        Inicializa o configurador com a configuração fornecida.
        
        Args:
            config: Configuração de observabilidade
            logger_name: Nome base do logger
        """
        self.config = config
        self.logger_name = logger_name
        self.logger = self._setup_logger()
        self.tracer = self._setup_tracer() if config.tracing_enabled else None
        self.meter = self._setup_meter() if config.metrics_enabled else None
    
    def _setup_logger(self) -> logging.Logger:
        """
        Configura o logger de acordo com as configurações.
        
        Returns:
            logging.Logger: Logger configurado
        """
        logger = logging.getLogger(self.logger_name)
        
        # Configurar nível de log
        log_level = getattr(logging, self.config.log_level, logging.INFO)
        logger.setLevel(log_level)
        
        # Limpar handlers existentes
        for handler in logger.handlers[:]:
            logger.removeHandler(handler)
        
        # Adicionar handler para console
        console_handler = logging.StreamHandler()
        console_handler.setLevel(log_level)
        
        # Configurar formato de log
        if self.config.log_format.lower() == "json":
            # Formatter JSON personalizado
            class JsonFormatter(logging.Formatter):
                def format(self, record):
                    timestamp = datetime.utcnow().isoformat() + "Z"
                    
                    log_data = {
                        "timestamp": timestamp,
                        "level": record.levelname,
                        "message": record.getMessage(),
                        "logger": record.name,
                        "service": self.config.service_name,
                        "environment": self.config.environment,
                        "tenant_id": self.config.tenant_id,
                        "module": self.config.module_name,
                    }
                    
                    # Adicionar componente se disponível
                    if self.config.component_name:
                        log_data["component"] = self.config.component_name
                    
                    # Adicionar região se disponível
                    if self.config.region:
                        log_data["region"] = self.config.region
                    
                    # Adicionar contexto de mercado se disponível
                    if self.config.market_context:
                        log_data["market_context"] = self.config.market_context
                    
                    # Adicionar informações de exceção se disponíveis
                    if record.exc_info:
                        log_data["exception"] = {
                            "type": record.exc_info[0].__name__,
                            "message": str(record.exc_info[1]),
                            "traceback": self.formatException(record.exc_info)
                        }
                    
                    # Adicionar atributos extras do record
                    for key, value in record.__dict__.items():
                        if key not in ['args', 'asctime', 'created', 'exc_info', 'exc_text', 'filename',
                                      'funcName', 'id', 'levelname', 'levelno', 'lineno', 'module',
                                      'msecs', 'message', 'msg', 'name', 'pathname', 'process',
                                      'processName', 'relativeCreated', 'stack_info', 'thread', 'threadName']:
                            log_data[key] = value
                    
                    return json.dumps(log_data)
            
            formatter = JsonFormatter(self.config)
        else:
            # Formatter padrão para output legível
            formatter = logging.Formatter(
                f'%(asctime)s [%(levelname)s] {self.config.service_name}::{self.config.tenant_id} - %(name)s - %(message)s'
            )
        
        console_handler.setFormatter(formatter)
        logger.addHandler(console_handler)
        
        # Adicionar handler para arquivo se configurado
        if self.config.log_file:
            file_handler = logging.FileHandler(self.config.log_file)
            file_handler.setLevel(log_level)
            file_handler.setFormatter(formatter)
            logger.addHandler(file_handler)
        
        # Adicionar filtro de tenant se habilitado
        if self.config.tenant_filtered_logging:
            class TenantFilter(logging.Filter):
                def __init__(self, tenant_id):
                    super().__init__()
                    self.tenant_id = tenant_id
                
                def filter(self, record):
                    # Adicionar tenant_id ao record se não estiver presente
                    if not hasattr(record, 'tenant_id'):
                        record.tenant_id = self.tenant_id
                    return True
            
            tenant_filter = TenantFilter(self.config.tenant_id)
            logger.addFilter(tenant_filter)
        
        return logger
    
    def _setup_tracer(self):
        """
        Configura o tracer OpenTelemetry.
        
        Returns:
            Tracer: Tracer configurado
        """
        # Esta implementação será expandida com a integração real do OpenTelemetry
        # Por agora, retorna um mock para funcionalidade básica
        class MockTracer:
            def __init__(self, config):
                self.config = config
            
            def start_span(self, name, **kwargs):
                class MockSpan:
                    def __init__(self, name, **kwargs):
                        self.name = name
                        self.attributes = kwargs.get("attributes", {})
                    
                    def __enter__(self):
                        return self
                    
                    def __exit__(self, exc_type, exc_val, exc_tb):
                        pass
                    
                    def set_attribute(self, key, value):
                        self.attributes[key] = value
                    
                    def record_exception(self, exception):
                        pass
                
                return MockSpan(name, **kwargs)
            
            def get_current_span(self):
                return None
        
        return MockTracer(self.config)
    
    def _setup_meter(self):
        """
        Configura o meter OpenTelemetry.
        
        Returns:
            Meter: Meter configurado
        """
        # Esta implementação será expandida com a integração real do OpenTelemetry
        # Por agora, retorna um mock para funcionalidade básica
        class MockMeter:
            def __init__(self, config):
                self.config = config
                self.metrics = {}
            
            def create_counter(self, name, description=None, unit=None):
                class MockCounter:
                    def __init__(self, name):
                        self.name = name
                        self.value = 0
                    
                    def add(self, amount, attributes=None):
                        self.value += amount
                
                counter = MockCounter(name)
                self.metrics[name] = counter
                return counter
            
            def create_histogram(self, name, description=None, unit=None):
                class MockHistogram:
                    def __init__(self, name):
                        self.name = name
                        self.values = []
                    
                    def record(self, value, attributes=None):
                        self.values.append(value)
                
                histogram = MockHistogram(name)
                self.metrics[name] = histogram
                return histogram
        
        return MockMeter(self.config)
    
    def get_layer_config(self, layer_name: str) -> Dict[str, Any]:
        """
        Obtém configuração específica para uma camada.
        
        Args:
            layer_name: Nome da camada (infrastructure, application, business, user, security)
        
        Returns:
            Dict[str, Any]: Configuração da camada
        """
        level_attr = f"{layer_name}_layer_level"
        if not hasattr(self.config, level_attr):
            raise ValueError(f"Camada inválida: {layer_name}")
        
        level = getattr(self.config, level_attr)
        
        # Configurações base para todas as camadas
        base_config = {
            "tenant_id": self.config.tenant_id,
            "service_name": self.config.service_name,
            "environment": self.config.environment,
            "module_name": self.config.module_name,
            "component_name": self.config.component_name,
            "observability_level": level,
        }
        
        # Ajustar configurações com base no nível de observabilidade
        if level == ObservabilityLevel.BASIC:
            base_config.update({
                "log_level": "WARNING",
                "tracing_enabled": False,
                "metrics_enabled": True,
                "metrics_export_interval_ms": 300000,  # 5 minutos
                "tracing_sample_rate": 0.01,
            })
        elif level == ObservabilityLevel.STANDARD:
            base_config.update({
                "log_level": "INFO",
                "tracing_enabled": True,
                "metrics_enabled": True,
                "metrics_export_interval_ms": 60000,  # 1 minuto
                "tracing_sample_rate": 0.1,
            })
        elif level == ObservabilityLevel.ADVANCED:
            base_config.update({
                "log_level": "INFO",
                "tracing_enabled": True,
                "metrics_enabled": True,
                "metrics_export_interval_ms": 30000,  # 30 segundos
                "tracing_sample_rate": 0.5,
                "high_cardinality_metrics": True,
            })
        elif level == ObservabilityLevel.DIAGNOSTIC:
            base_config.update({
                "log_level": "DEBUG",
                "tracing_enabled": True,
                "metrics_enabled": True,
                "metrics_export_interval_ms": 15000,  # 15 segundos
                "tracing_sample_rate": 1.0,
                "high_cardinality_metrics": True,
            })
        
        # Camada-específica: ajustes adicionais para segurança
        if layer_name == "security":
            base_config.update({
                "pii_masking_enabled": True,
                "compliance_tracking_enabled": True,
            })
        
        return base_config
    
    @classmethod
    def from_env(cls, logger_name: str = "innovabiz.observability") -> 'MultiLayerObservabilityConfigurator':
        """
        Cria um configurador a partir de variáveis de ambiente.
        
        Args:
            logger_name: Nome do logger
        
        Returns:
            MultiLayerObservabilityConfigurator: Configurador inicializado
        """
        config = MultiLayerObservabilityConfig.from_env()
        return cls(config, logger_name)