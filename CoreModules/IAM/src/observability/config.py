"""
Módulo de configuração para observabilidade do IAM Audit Service.

Este módulo define as classes de configuração para os diferentes componentes
de observabilidade, incluindo métricas, health checks e rastreamento.

Fornece suporte para configuração via environment variables, com valores
padrão sensíveis e validação completa.
"""

import os
from typing import Dict, List, Optional, Union, Any
from pydantic import BaseSettings, Field, validator


class MetricsConfig(BaseSettings):
    """Configuração para coleta e exposição de métricas."""
    
    enabled: bool = True
    namespace: str = "innovabiz"
    subsystem: str = "iam_audit"
    
    # Configurações de métricas HTTP
    http_request_duration_buckets: List[float] = [
        0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0
    ]
    
    # Configurações de métricas de eventos de auditoria
    audit_event_processing_buckets: List[float] = [
        0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5
    ]
    
    # Configurações de métricas de compliance
    compliance_check_buckets: List[float] = [
        0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0
    ]
    
    # Configurações de métricas de retenção
    retention_execution_buckets: List[float] = [
        0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0
    ]
    
    # Configurações de exposição de métricas
    exclude_paths: List[str] = ["/metrics", "/health", "/live", "/ready", "/diagnostic"]
    
    class Config:
        env_prefix = "METRICS_"
        env_nested_delimiter = "__"


class HealthConfig(BaseSettings):
    """Configuração para health checks e diagnósticos."""
    
    enabled: bool = True
    
    # Configurações de timeout para verificações (ms)
    database_timeout_ms: int = 500
    cache_timeout_ms: int = 200
    kafka_timeout_ms: int = 300
    storage_timeout_ms: int = 400
    
    # Componentes a verificar em cada tipo de health check
    liveness_components: List[str] = ["process"]
    readiness_components: List[str] = ["database", "cache", "kafka", "storage"]
    health_components: List[str] = ["database", "cache"]
    
    # Configurações para diagnósticos
    enable_diagnostic: bool = True
    diagnostic_requires_auth: bool = True
    
    class Config:
        env_prefix = "HEALTH_"
        env_nested_delimiter = "__"


class TracingConfig(BaseSettings):
    """Configuração para rastreamento distribuído."""
    
    enabled: bool = True
    service_name: str = "iam-audit-service"
    namespace: str = "innovabiz"
    
    # Configuração do exportador OTLP
    otlp_endpoint: str = "http://otel-collector:4317"
    otlp_headers: Dict[str, str] = {}
    
    # URLs a serem excluídas da instrumentação automática
    excluded_urls: List[str] = ["/health", "/live", "/metrics", "/ready", "/diagnostic"]
    
    # Configuração de sampling
    sample_ratio: float = 1.0  # 100% por padrão
    
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


class AlertingConfig(BaseSettings):
    """Configuração para alertas."""
    
    enabled: bool = True
    alertmanager_url: str = "http://alertmanager:9093/api/v2/alerts"
    
    # Canais de notificação
    notification_channels: Dict[str, Dict[str, str]] = {
        "slack": {"webhook": ""},
        "email": {"recipients": ""},
        "pagerduty": {"service_key": ""}
    }
    
    # Severidades e canais associados
    severity_channels: Dict[str, List[str]] = {
        "critical": ["slack", "email", "pagerduty"],
        "warning": ["slack", "email"],
        "info": ["slack"]
    }
    
    class Config:
        env_prefix = "ALERTING_"
        env_nested_delimiter = "__"


class ObservabilityConfig(BaseSettings):
    """Configuração principal para observabilidade."""
    
    service_name: str = "iam-audit-service"
    service_version: str = "1.0.0"
    environment: str = "production"
    
    # Configurações de contexto padrão
    default_tenant: str = "default"
    default_region: str = "global"
    
    # Labels padrão para todas as métricas
    default_labels: Dict[str, str] = {
        "service": "iam-audit-service",
        "component": "audit"
    }
    
    # Prefixo para rotas de observabilidade
    route_prefix: str = ""
    
    # Subconfiguração para cada componente
    metrics: MetricsConfig = Field(default_factory=MetricsConfig)
    health: HealthConfig = Field(default_factory=HealthConfig)
    tracing: TracingConfig = Field(default_factory=TracingConfig)
    alerting: AlertingConfig = Field(default_factory=AlertingConfig)
    
    class Config:
        env_prefix = "OBSERVABILITY_"
        env_nested_delimiter = "__"
    
    @validator("default_labels", pre=True)
    def set_version_in_labels(cls, v, values):
        """Adiciona automaticamente a versão do serviço às labels padrão."""
        if isinstance(v, dict) and "version" not in v and "service_version" in values:
            v["version"] = values["service_version"]
        return v


def configure_observability(app: Any, **kwargs) -> "ObservabilityIntegration":
    """
    Configura e inicializa observabilidade para uma aplicação FastAPI.
    
    Args:
        app: Aplicação FastAPI a ser instrumentada
        **kwargs: Parâmetros adicionais para sobrescrever configurações padrão
    
    Returns:
        Uma instância configurada de ObservabilityIntegration
    """
    # Importação tardia para evitar referência circular
    from .integration import ObservabilityIntegration
    
    # Cria configuração com base nas variáveis de ambiente e parâmetros
    config = ObservabilityConfig(**kwargs)
    
    # Inicializa a integração
    obs = ObservabilityIntegration(config=config)
    
    # Instrumenta a aplicação
    obs.instrument_app(app)
    
    return obs