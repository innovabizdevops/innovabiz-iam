"""
Módulo de métricas para o IAM Audit Service.

Implementa o gerenciamento de métricas Prometheus, incluindo decoradores
para instrumentação automática, factory methods para criação de métricas,
e funções utilitárias.

Métricas seguem os padrões estabelecidos no ADR-004, com suporte para
multi-tenant, multi-região e categorias padronizadas de métricas.
"""

import time
import functools
import inspect
from typing import Dict, List, Optional, Callable, Any, Type, Union
from datetime import datetime
from enum import Enum

import prometheus_client
from prometheus_client import Counter, Histogram, Gauge, Summary, CollectorRegistry
from fastapi import Request, Response

from .config import MetricsConfig


class MetricCategory(str, Enum):
    """Categorias padrão para métricas do IAM Audit Service."""
    
    HTTP = "http"
    AUDIT = "audit"
    COMPLIANCE = "compliance"
    RETENTION = "retention"
    RESOURCE = "resource"
    HEALTH = "health"
    BUSINESS = "business"


class MetricsManager:
    """
    Gerenciador de métricas para o IAM Audit Service.
    
    Responsável por criar e gerenciar métricas Prometheus, seguindo
    os padrões estabelecidos para o IAM Audit Service.
    """
    
    def __init__(
        self,
        registry: Optional[CollectorRegistry] = None,
        config: Optional[MetricsConfig] = None,
        default_labels: Optional[Dict[str, str]] = None
    ):
        """
        Inicializa o gerenciador de métricas.
        
        Args:
            registry: Registry Prometheus para registrar métricas
            config: Configuração para métricas
            default_labels: Labels padrão para todas as métricas
        """
        self.registry = registry or prometheus_client.REGISTRY
        self.config = config or MetricsConfig()
        self.default_labels = default_labels or {}
        
        # Inicializa métricas padrão
        self._setup_default_metrics()
    
    def _setup_default_metrics(self):
        """Configura as métricas padrão para o serviço."""
        
        # Métricas de serviço
        self.service_start_counter = self.create_counter(
            name="service_start_total",
            description="Total de inicializações do serviço",
            labelnames=[]
        )
        
        self.service_info = self.create_gauge(
            name="service_info",
            description="Informações sobre o serviço",
            labelnames=["version", "build_id"]
        )
        
        # Métricas HTTP (padrão RED: Request Rate, Error Rate, Duration)
        self.http_requests_total = self.create_counter(
            name="http_request_total",
            description="Total de requisições HTTP",
            category=MetricCategory.HTTP,
            labelnames=["method", "path", "tenant", "region"]
        )
        
        self.http_responses_total = self.create_counter(
            name="http_response_total",
            description="Total de respostas HTTP por código de status",
            category=MetricCategory.HTTP,
            labelnames=["method", "path", "status", "tenant", "region"]
        )
        
        self.http_exceptions_total = self.create_counter(
            name="http_exception_total",
            description="Total de exceções em requisições HTTP",
            category=MetricCategory.HTTP,
            labelnames=["method", "path", "exception", "tenant", "region"]
        )
        
        self.http_request_duration_seconds = self.create_histogram(
            name="http_request_duration_seconds",
            description="Duração das requisições HTTP em segundos",
            category=MetricCategory.HTTP,
            labelnames=["method", "path", "tenant", "region"],
            buckets=self.config.http_request_duration_buckets
        )
        
        # Métricas de eventos de auditoria
        self.audit_event_processed_total = self.create_counter(
            name="event_processed_total",
            description="Total de eventos de auditoria processados",
            category=MetricCategory.AUDIT,
            labelnames=["event_type", "status", "tenant", "region"]
        )
        
        self.audit_event_processing_seconds = self.create_histogram(
            name="event_processing_seconds",
            description="Duração do processamento de eventos de auditoria",
            category=MetricCategory.AUDIT,
            labelnames=["event_type", "tenant", "region"],
            buckets=self.config.audit_event_processing_buckets
        )
        
        # Métricas de compliance
        self.compliance_check_total = self.create_counter(
            name="compliance_check_total",
            description="Total de verificações de compliance realizadas",
            category=MetricCategory.COMPLIANCE,
            labelnames=["compliance_type", "status", "tenant", "region"]
        )
        
        self.compliance_violation_total = self.create_counter(
            name="compliance_violation_total",
            description="Total de violações de compliance detectadas",
            category=MetricCategory.COMPLIANCE,
            labelnames=["compliance_type", "severity", "tenant", "region"]
        )
        
        self.compliance_check_seconds = self.create_histogram(
            name="compliance_check_seconds",
            description="Duração das verificações de compliance",
            category=MetricCategory.COMPLIANCE,
            labelnames=["compliance_type", "tenant", "region"],
            buckets=self.config.compliance_check_buckets
        )
        
        # Métricas de retenção
        self.retention_execution_total = self.create_counter(
            name="retention_execution_total",
            description="Total de execuções de políticas de retenção",
            category=MetricCategory.RETENTION,
            labelnames=["retention_policy", "status", "tenant", "region"]
        )
        
        self.retention_purge_total = self.create_counter(
            name="retention_purge_total",
            description="Total de registros expurgados por política de retenção",
            category=MetricCategory.RETENTION,
            labelnames=["retention_policy", "tenant", "region"]
        )
        
        self.retention_execution_seconds = self.create_histogram(
            name="retention_execution_seconds",
            description="Duração das execuções de políticas de retenção",
            category=MetricCategory.RETENTION,
            labelnames=["retention_policy", "tenant", "region"],
            buckets=self.config.retention_execution_buckets
        )
        
        # Métricas de recursos
        self.resource_utilization_ratio = self.create_gauge(
            name="resource_utilization_ratio",
            description="Taxa de utilização de recursos (0-1)",
            category=MetricCategory.RESOURCE,
            labelnames=["resource_type", "tenant", "region"]
        )
        
        # Métricas de health checks
        self.health_check_total = self.create_counter(
            name="health_check_total",
            description="Total de health checks realizados",
            category=MetricCategory.HEALTH,
            labelnames=["check_type", "status", "tenant", "region"]
        )
        
        self.dependency_check_failed = self.create_counter(
            name="dependency_check_failed",
            description="Total de falhas em verificações de dependências",
            category=MetricCategory.HEALTH,
            labelnames=["dependency", "tenant", "region"]
        )
    
    def create_counter(
        self,
        name: str,
        description: str,
        labelnames: List[str],
        category: Optional[Union[str, MetricCategory]] = None,
        registry: Optional[CollectorRegistry] = None
    ) -> Counter:
        """
        Cria um contador Prometheus com o nome e descrição fornecidos.
        
        Args:
            name: Nome da métrica
            description: Descrição da métrica
            labelnames: Nomes das labels para a métrica
            category: Categoria da métrica (opcional)
            registry: Registry Prometheus personalizado (opcional)
        
        Returns:
            Um contador Prometheus configurado
        """
        # Aplica prefixo de namespace e subsystem
        metric_name = self._get_metric_name(name, category)
        
        # Cria o contador
        return Counter(
            metric_name,
            description,
            labelnames=labelnames,
            registry=registry or self.registry
        )
    
    def create_gauge(
        self,
        name: str,
        description: str,
        labelnames: List[str],
        category: Optional[Union[str, MetricCategory]] = None,
        registry: Optional[CollectorRegistry] = None
    ) -> Gauge:
        """
        Cria um gauge Prometheus com o nome e descrição fornecidos.
        
        Args:
            name: Nome da métrica
            description: Descrição da métrica
            labelnames: Nomes das labels para a métrica
            category: Categoria da métrica (opcional)
            registry: Registry Prometheus personalizado (opcional)
        
        Returns:
            Um gauge Prometheus configurado
        """
        # Aplica prefixo de namespace e subsystem
        metric_name = self._get_metric_name(name, category)
        
        # Cria o gauge
        return Gauge(
            metric_name,
            description,
            labelnames=labelnames,
            registry=registry or self.registry
        )
    
    def create_histogram(
        self,
        name: str,
        description: str,
        labelnames: List[str],
        buckets: Optional[List[float]] = None,
        category: Optional[Union[str, MetricCategory]] = None,
        registry: Optional[CollectorRegistry] = None
    ) -> Histogram:
        """
        Cria um histograma Prometheus com o nome e descrição fornecidos.
        
        Args:
            name: Nome da métrica
            description: Descrição da métrica
            labelnames: Nomes das labels para a métrica
            buckets: Buckets personalizados para o histograma
            category: Categoria da métrica (opcional)
            registry: Registry Prometheus personalizado (opcional)
        
        Returns:
            Um histograma Prometheus configurado
        """
        # Aplica prefixo de namespace e subsystem
        metric_name = self._get_metric_name(name, category)
        
        # Cria o histograma
        return Histogram(
            metric_name,
            description,
            labelnames=labelnames,
            buckets=buckets,
            registry=registry or self.registry
        )
    
    def create_summary(
        self,
        name: str,
        description: str,
        labelnames: List[str],
        quantiles: Optional[Dict[float, float]] = None,
        category: Optional[Union[str, MetricCategory]] = None,
        registry: Optional[CollectorRegistry] = None
    ) -> Summary:
        """
        Cria um summary Prometheus com o nome e descrição fornecidos.
        
        Args:
            name: Nome da métrica
            description: Descrição da métrica
            labelnames: Nomes das labels para a métrica
            quantiles: Quantis personalizados para o summary
            category: Categoria da métrica (opcional)
            registry: Registry Prometheus personalizado (opcional)
        
        Returns:
            Um summary Prometheus configurado
        """
        # Aplica prefixo de namespace e subsystem
        metric_name = self._get_metric_name(name, category)
        
        # Cria o summary
        return Summary(
            metric_name,
            description,
            labelnames=labelnames,
            registry=registry or self.registry
        )
    
    def _get_metric_name(self, name: str, category: Optional[Union[str, MetricCategory]] = None) -> str:
        """
        Obtém o nome completo da métrica com namespace, subsystem e categoria.
        
        Args:
            name: Nome base da métrica
            category: Categoria da métrica (opcional)
        
        Returns:
            Nome completo da métrica formatado como namespace_subsystem_category_name
        """
        parts = []
        
        if self.config.namespace:
            parts.append(self.config.namespace)
        
        if self.config.subsystem:
            parts.append(self.config.subsystem)
        
        if category:
            if isinstance(category, MetricCategory):
                parts.append(category.value)
            else:
                parts.append(category)
        
        parts.append(name)
        
        return "_".join(parts)


def instrument_function(
    histogram: Optional[Histogram] = None,
    counter: Optional[Counter] = None,
    labels: Optional[Dict[str, str]] = None,
    extract_labels_from_args: Optional[Dict[str, str]] = None
):
    """
    Decorador para instrumentar uma função com métricas Prometheus.
    
    Args:
        histogram: Histograma Prometheus para medir duração
        counter: Contador Prometheus para incrementar a cada chamada
        labels: Labels estáticas para as métricas
        extract_labels_from_args: Mapeamento de nomes de argumentos para labels
    
    Returns:
        Decorador configurado
    """
    def decorator(func):
        @functools.wraps(func)
        async def async_wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Prepara labels
            final_labels = labels.copy() if labels else {}
            
            # Extrai labels dos argumentos
            if extract_labels_from_args:
                for label_name, arg_name in extract_labels_from_args.items():
                    if arg_name in kwargs:
                        final_labels[label_name] = kwargs[arg_name]
            
            try:
                # Executa a função original
                result = await func(*args, **kwargs)
                
                # Registra métricas de sucesso
                duration = time.time() - start_time
                
                if histogram:
                    histogram.labels(**final_labels).observe(duration)
                
                if counter:
                    counter.labels(**final_labels).inc()
                
                return result
                
            except Exception as exc:
                # Registra métricas de erro
                duration = time.time() - start_time
                
                if histogram:
                    histogram.labels(**final_labels).observe(duration)
                
                # Re-lança a exceção
                raise
        
        @functools.wraps(func)
        def sync_wrapper(*args, **kwargs):
            start_time = time.time()
            
            # Prepara labels
            final_labels = labels.copy() if labels else {}
            
            # Extrai labels dos argumentos
            if extract_labels_from_args:
                for label_name, arg_name in extract_labels_from_args.items():
                    if arg_name in kwargs:
                        final_labels[label_name] = kwargs[arg_name]
            
            try:
                # Executa a função original
                result = func(*args, **kwargs)
                
                # Registra métricas de sucesso
                duration = time.time() - start_time
                
                if histogram:
                    histogram.labels(**final_labels).observe(duration)
                
                if counter:
                    counter.labels(**final_labels).inc()
                
                return result
                
            except Exception as exc:
                # Registra métricas de erro
                duration = time.time() - start_time
                
                if histogram:
                    histogram.labels(**final_labels).observe(duration)
                
                # Re-lança a exceção
                raise
        
        # Verifica se a função é assíncrona ou síncrona
        if inspect.iscoroutinefunction(func):
            return async_wrapper
        return sync_wrapper
    
    return decorator


def instrument_audit_event(
    event_type: str = "generic",
    metrics_manager: Optional[MetricsManager] = None
):
    """
    Decorador para instrumentar funções que processam eventos de auditoria.
    
    Args:
        event_type: Tipo do evento de auditoria
        metrics_manager: Gerenciador de métricas a ser usado
    
    Returns:
        Decorador configurado
    """
    def decorator(func):
        @functools.wraps(func)
        async def async_wrapper(*args, **kwargs):
            # Obtém o metrics_manager correto
            mm = metrics_manager
            if not mm:
                # Tenta obter do primeiro argumento (assumindo ser self)
                if args and hasattr(args[0], "metrics"):
                    mm = args[0].metrics
            
            if not mm:
                # Fallback para uma instância global/singleton
                from .integration import get_metrics_manager
                mm = get_metrics_manager()
            
            # Extrai tenant e region do request ou kwargs
            tenant = "default"
            region = "global"
            
            if "tenant" in kwargs:
                tenant = kwargs["tenant"]
            
            if "region" in kwargs:
                region = kwargs["region"]
            
            # Prepara labels
            labels = {
                "event_type": event_type,
                "tenant": tenant,
                "region": region
            }
            
            start_time = time.time()
            
            try:
                # Executa a função original
                result = await func(*args, **kwargs)
                
                # Registra métricas de sucesso
                duration = time.time() - start_time
                
                mm.audit_event_processed_total.labels(
                    event_type=event_type,
                    status="success",
                    tenant=tenant,
                    region=region
                ).inc()
                
                mm.audit_event_processing_seconds.labels(
                    event_type=event_type,
                    tenant=tenant,
                    region=region
                ).observe(duration)
                
                return result
                
            except Exception as exc:
                # Registra métricas de erro
                duration = time.time() - start_time
                
                mm.audit_event_processed_total.labels(
                    event_type=event_type,
                    status="error",
                    tenant=tenant,
                    region=region
                ).inc()
                
                mm.audit_event_processing_seconds.labels(
                    event_type=event_type,
                    tenant=tenant,
                    region=region
                ).observe(duration)
                
                # Re-lança a exceção
                raise
        
        @functools.wraps(func)
        def sync_wrapper(*args, **kwargs):
            # Implementação similar para funções síncronas
            # ...
            pass
        
        # Verifica se a função é assíncrona ou síncrona
        if inspect.iscoroutinefunction(func):
            return async_wrapper
        return sync_wrapper
    
    return decorator


# Singleton global para acesso em outras partes do código
_metrics_manager_instance = None

def get_metrics_manager() -> MetricsManager:
    """
    Obtém a instância singleton do gerenciador de métricas.
    
    Returns:
        Instância do MetricsManager
    """
    global _metrics_manager_instance
    
    if _metrics_manager_instance is None:
        _metrics_manager_instance = MetricsManager()
    
    return _metrics_manager_instance


def set_metrics_manager(manager: MetricsManager) -> None:
    """
    Define a instância singleton do gerenciador de métricas.
    
    Args:
        manager: Instância do MetricsManager a ser usada como singleton
    """
    global _metrics_manager_instance
    _metrics_manager_instance = manager