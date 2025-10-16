"""
Configuração de observabilidade para o serviço IAM Audit.

Este módulo fornece configurações e utilitários para integrar facilmente
o módulo de observabilidade com a aplicação principal FastAPI. Inclui
configurações para ambientes de produção, staging e desenvolvimento,
além de suporte completo para multi-tenant e multi-regional.

Author: INNOVABIZ DevOps Team
Date: 2025-07-31
"""

import os
from functools import lru_cache
from typing import Dict, List, Optional

from pydantic import BaseModel, Field, validator

from api.app.integrations.observability import setup_observability


class AlertConfig(BaseModel):
    """Configuração de alertas para o serviço."""
    alert_manager_url: str = Field(
        default="http://alertmanager:9093",
        description="URL do Alertmanager para envio de alertas"
    )
    severity_levels: Dict[str, int] = Field(
        default={
            "critical": 1,
            "high": 2,
            "medium": 3,
            "low": 4,
            "info": 5
        },
        description="Níveis de severidade para alertas"
    )
    notification_channels: Dict[str, List[str]] = Field(
        default={
            "critical": ["slack", "email", "sms"],
            "high": ["slack", "email"],
            "medium": ["slack"],
            "low": ["slack"],
            "info": ["slack"]
        },
        description="Canais de notificação por nível de severidade"
    )


class RegionalConfig(BaseModel):
    """Configuração específica de região."""
    region_code: str
    prometheus_url: str = Field(
        default="http://prometheus:9090",
        description="URL do Prometheus para scraping de métricas"
    )
    scrape_interval_seconds: int = Field(
        default=15,
        description="Intervalo de coleta de métricas em segundos"
    )
    compliance_frameworks: List[str] = Field(
        default_factory=list,
        description="Frameworks de compliance aplicáveis à região"
    )
    
    @validator("compliance_frameworks", pre=True)
    def set_default_compliance_frameworks(cls, v, values):
        """Define frameworks de compliance padrão baseados na região."""
        if not v:
            region = values.get("region_code", "").lower()
            if region.startswith("br"):
                return ["LGPD", "SOX", "PCI-DSS", "ISO-27001"]
            elif region.startswith("eu"):
                return ["GDPR", "SOX", "PCI-DSS", "ISO-27001"]
            elif region.startswith("us"):
                return ["CCPA", "SOX", "PCI-DSS", "ISO-27001"]
            elif region.startswith("ao"):
                return ["BNA-REGULATION", "SOX", "PCI-DSS", "ISO-27001"]
            else:
                return ["PCI-DSS", "ISO-27001"]
        return v


class ObservabilityConfig(BaseModel):
    """Configuração de observabilidade para o serviço IAM Audit."""
    enabled: bool = Field(
        default=True,
        description="Habilita ou desabilita a observabilidade"
    )
    service_name: str = Field(
        default="iam-audit",
        description="Nome do serviço para registro em métricas"
    )
    service_version: str = Field(
        default="1.0.0",
        description="Versão do serviço para registro em métricas"
    )
    environment: str = Field(
        default="production",
        description="Ambiente de execução (production, staging, development)"
    )
    default_region: str = Field(
        default="br-east",
        description="Região padrão para registro em métricas"
    )
    metrics_endpoint: str = Field(
        default="/metrics",
        description="Endpoint para exposição de métricas do Prometheus"
    )
    health_endpoint: str = Field(
        default="/health",
        description="Endpoint para verificação de saúde do serviço"
    )
    diagnostic_endpoint: str = Field(
        default="/diagnostic",
        description="Endpoint para diagnóstico do serviço"
    )
    header_tenant_id: str = Field(
        default="X-Tenant-ID",
        description="Header HTTP para identificação do tenant"
    )
    header_region: str = Field(
        default="X-Region",
        description="Header HTTP para identificação da região"
    )
    header_correlation_id: str = Field(
        default="X-Correlation-ID",
        description="Header HTTP para identificação da correlação"
    )
    regions: Dict[str, RegionalConfig] = Field(
        default_factory=dict,
        description="Configurações específicas por região"
    )
    alerting: AlertConfig = Field(
        default_factory=AlertConfig,
        description="Configuração de alertas"
    )
    log_metrics: bool = Field(
        default=True,
        description="Habilita ou desabilita o log de métricas"
    )
    tracing_enabled: bool = Field(
        default=True,
        description="Habilita ou desabilita o rastreamento distribuído"
    )
    tracing_sample_rate: float = Field(
        default=0.1,
        description="Taxa de amostragem para rastreamento (0.0 a 1.0)"
    )
    
    @validator("regions", pre=True)
    def set_default_regions(cls, v, values):
        """Define regiões padrão se nenhuma for fornecida."""
        if not v:
            default_regions = {
                "br-east": RegionalConfig(
                    region_code="br-east",
                    prometheus_url="http://prometheus-br-east:9090"
                ),
                "br-south": RegionalConfig(
                    region_code="br-south",
                    prometheus_url="http://prometheus-br-south:9090"
                ),
                "us-east": RegionalConfig(
                    region_code="us-east",
                    prometheus_url="http://prometheus-us-east:9090"
                ),
                "eu-central": RegionalConfig(
                    region_code="eu-central",
                    prometheus_url="http://prometheus-eu-central:9090"
                ),
                "ao-central": RegionalConfig(
                    region_code="ao-central",
                    prometheus_url="http://prometheus-ao-central:9090"
                )
            }
            return default_regions
        return v


@lru_cache()
def get_observability_config() -> ObservabilityConfig:
    """
    Retorna a configuração de observabilidade baseada em variáveis de ambiente.
    
    Prioriza variáveis de ambiente, depois valores de configuração padrão.
    Usa cache para evitar recarregar a configuração a cada chamada.
    """
    # Obter valores de variáveis de ambiente ou usar padrões
    service_version = os.getenv("SERVICE_VERSION", "1.0.0")
    build_id = os.getenv("BUILD_ID", "local")
    commit_hash = os.getenv("COMMIT_HASH", "unknown")
    environment = os.getenv("ENVIRONMENT", "development")
    default_region = os.getenv("DEFAULT_REGION", "br-east")
    
    # Criar configuração base
    config = ObservabilityConfig(
        service_version=service_version,
        environment=environment,
        default_region=default_region
    )
    
    # Retornar configuração
    return config


def configure_observability(app, db_client=None, redis_client=None, kafka_client=None, storage_client=None):
    """
    Configura observabilidade para a aplicação FastAPI.
    
    Simplifica a integração da observabilidade com a aplicação principal,
    aplicando a configuração padrão ou personalizada.
    
    Args:
        app: Instância da aplicação FastAPI
        db_client: Cliente de banco de dados para health checks
        redis_client: Cliente Redis para health checks
        kafka_client: Cliente Kafka para health checks
        storage_client: Cliente de armazenamento para health checks
    
    Returns:
        A instância da aplicação FastAPI com observabilidade configurada
    """
    # Obter configuração
    config = get_observability_config()
    
    # Se observabilidade estiver desabilitada, retornar app sem modificações
    if not config.enabled:
        return app
    
    # Obter variáveis de ambiente adicionais
    build_id = os.getenv("BUILD_ID", "local")
    commit_hash = os.getenv("COMMIT_HASH", "unknown")
    
    # Configurar observabilidade
    setup_observability(
        app=app,
        service_name=config.service_name,
        service_version=config.service_version,
        build_id=build_id,
        commit_hash=commit_hash,
        environment=config.environment,
        region=config.default_region,
        db_client=db_client,
        redis_client=redis_client,
        kafka_client=kafka_client,
        storage_client=storage_client
    )
    
    return app