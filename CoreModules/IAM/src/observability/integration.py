"""
Módulo principal de integração de observabilidade para o IAM Audit Service.

Este módulo implementa a classe ObservabilityIntegration que serve como ponto central
para a integração de métricas, health checks, e rastreamento distribuído no serviço.

Referências:
- ADR-005: Framework de Decoradores e Middleware
- ADR-006: Health Checks e Diagnósticos
- ADR-007: Sistema de Alertas e Dashboards
- ADR-008: Rastreamento com OpenTelemetry
"""

import time
import logging
from typing import Dict, List, Optional, Any, Callable, Union
from datetime import datetime

import prometheus_client
from prometheus_client import Counter, Histogram, Gauge, Summary, CollectorRegistry
from prometheus_client.exposition import CONTENT_TYPE_LATEST
from fastapi import FastAPI, APIRouter, Request, Response, Depends, Header
from fastapi.middleware.base import BaseHTTPMiddleware
from pydantic import BaseModel

from .config import ObservabilityConfig
from .metrics import MetricsManager, get_metrics_manager, set_metrics_manager
from .health import HealthChecker, HealthResponse, DiagnosticResponse
from .tracing import TracingIntegration

logger = logging.getLogger(__name__)


class ObservabilityIntegration:
    """
    Classe principal para integração de observabilidade no IAM Audit Service.
    
    Esta classe integra métricas Prometheus, health checks e rastreamento distribuído
    em uma solução unificada, com suporte a contextos multi-tenant e multi-regionais.
    """
    
    def __init__(
        self,
        config: ObservabilityConfig = None,
        registry: CollectorRegistry = None
    ):
        """
        Inicializa a integração de observabilidade.
        
        Args:
            config: Configuração para observabilidade. Se None, usa configuração padrão.
            registry: Registry Prometheus personalizado. Se None, usa o registry padrão.
        """
        self.service_start_time = datetime.utcnow()
        self.config = config or ObservabilityConfig()
        self.registry = registry or CollectorRegistry()
        
        # Inicializa os componentes
        self.metrics = MetricsManager(
            registry=self.registry,
            config=self.config.metrics,
            default_labels=self.config.default_labels
        )
        
        # Define o gerenciador de métricas como singleton global
        set_metrics_manager(self.metrics)
        
        self.health = HealthChecker(
            config=self.config.health,
            metrics=self.metrics,
            service_start_time=self.service_start_time
        )
        
        # Inicializa o tracer se a configuração de tracing estiver habilitada
        self.tracing = None
        if self.config.tracing and self.config.tracing.enabled:
            self.tracing = TracingIntegration(config=self.config.tracing)
            
        # Router para endpoints de observabilidade
        self.router = APIRouter(tags=["Observabilidade"])
        self._setup_routes()
        
        logger.info(
            f"ObservabilityIntegration inicializado: "
            f"metrics={self.config.metrics.enabled}, "
            f"health={self.config.health.enabled}, "
            f"tracing={bool(self.tracing)}"
        )
    
    def _setup_routes(self):
        """Configura as rotas do FastAPI para endpoints de observabilidade."""
        
        @self.router.get("/metrics", response_class=Response)
        async def metrics():
            """Expõe métricas no formato Prometheus."""
            return Response(
                prometheus_client.generate_latest(self.registry),
                media_type=CONTENT_TYPE_LATEST
            )
        
        @self.router.get("/health", response_model_exclude_none=True)
        async def health(
            request: Request,
            tenant: str = Header(None, alias="X-Tenant-ID"),
            region: str = Header(None, alias="X-Region")
        ):
            """
            Health check básico para verificar disponibilidade do serviço.
            Executa verificações rápidas (<100ms) para componentes essenciais.
            """
            tenant = tenant or self.config.default_tenant
            region = region or self.config.default_region
            
            result = await self.health.check_health(tenant, region)
            return result
        
        @self.router.get("/ready", response_model_exclude_none=True)
        async def ready(
            request: Request,
            tenant: str = Header(None, alias="X-Tenant-ID"),
            region: str = Header(None, alias="X-Region")
        ):
            """
            Readiness check para verificar capacidade de processar requisições.
            Usado por orquestradores para decisões de roteamento.
            """
            tenant = tenant or self.config.default_tenant
            region = region or self.config.default_region
            
            result = await self.health.check_readiness(tenant, region)
            return result
        
        @self.router.get("/live")
        async def live():
            """
            Liveness check simples para verificar se o processo está respondendo.
            Usado por orquestradores para detectar processos travados.
            """
            return {
                "status": "alive",
                "timestamp": datetime.utcnow().isoformat(),
            }
        
        @self.router.get("/diagnostic", response_model_exclude_none=True)
        async def diagnostic(
            request: Request,
            include_deps: bool = True,
            include_metrics: bool = True,
            include_config: bool = False,
            tenant: str = Header(None, alias="X-Tenant-ID"),
            region: str = Header(None, alias="X-Region"),
            # Em produção, adicionar: current_user: User = Depends(get_current_admin_user)
        ):
            """
            Diagnóstico detalhado do serviço com verificações profundas.
            Em produção, deve requerer autenticação de administrador.
            """
            tenant = tenant or self.config.default_tenant
            region = region or self.config.default_region
            
            result = await self.health.get_diagnostic(
                tenant=tenant,
                region=region,
                include_deps=include_deps,
                include_metrics=include_metrics,
                include_config=include_config
            )
            return result
    
    def instrument_app(self, app: FastAPI) -> None:
        """
        Instrumenta uma aplicação FastAPI com observabilidade completa.
        
        Args:
            app: Aplicação FastAPI a ser instrumentada
        """
        # Adiciona middleware para métricas HTTP
        if self.config.metrics.enabled:
            app.add_middleware(
                HTTPMetricsMiddleware,
                metrics_manager=self.metrics,
                exclude_paths=self.config.metrics.exclude_paths
            )
        
        # Adiciona middleware para propagação de contexto
        app.add_middleware(
            ContextMiddleware,
            default_tenant=self.config.default_tenant,
            default_region=self.config.default_region
        )
        
        # Configura instrumentação de tracing se habilitado
        if self.tracing:
            self.tracing.instrument_app(app)
        
        # Adiciona endpoints de observabilidade
        app.include_router(self.router, prefix=self.config.route_prefix)
        
        # Registra handlers para eventos do ciclo de vida
        @app.on_event("startup")
        async def startup_event():
            """Handler para evento de startup da aplicação."""
            await self._handle_startup()
        
        @app.on_event("shutdown")
        async def shutdown_event():
            """Handler para evento de shutdown da aplicação."""
            await self._handle_shutdown()
    
    async def _handle_startup(self) -> None:
        """
        Manipulador para evento de startup da aplicação.
        Inicializa recursos e registra métricas de inicialização.
        """
        logger.info("Inicializando recursos de observabilidade")
        
        # Incrementa contador de inicialização do serviço
        if self.metrics:
            self.metrics.service_start_counter.inc()
        
        # Inicializa verificadores de saúde
        await self.health.initialize()
        
        # Inicializa recursos de tracing
        if self.tracing:
            await self.tracing.initialize()
    
    async def _handle_shutdown(self) -> None:
        """
        Manipulador para evento de shutdown da aplicação.
        Realiza limpeza de recursos.
        """
        logger.info("Finalizando recursos de observabilidade")
        
        # Finaliza verificadores de saúde
        await self.health.shutdown()
        
        # Finaliza recursos de tracing
        if self.tracing:
            await self.tracing.shutdown()


class ContextInfo:
    """Classe auxiliar para armazenar informações de contexto."""
    
    def __init__(self, tenant: str, region: str, environment: str):
        self.tenant = tenant
        self.region = region
        self.environment = environment


class HTTPMetricsMiddleware(BaseHTTPMiddleware):
    """
    Middleware para instrumentação automática de métricas HTTP.
    
    Coleta métricas RED (Rate, Error, Duration) para todas as requisições HTTP,
    com suporte para contextos multi-tenant e multi-regionais.
    """
    
    def __init__(
        self,
        app: FastAPI,
        metrics_manager: MetricsManager,
        exclude_paths: List[str] = None
    ):
        """
        Inicializa o middleware de métricas HTTP.
        
        Args:
            app: Aplicação FastAPI
            metrics_manager: Gerenciador de métricas
            exclude_paths: Caminhos a serem excluídos da instrumentação
        """
        super().__init__(app)
        self.metrics = metrics_manager
        self.exclude_paths = exclude_paths or ["/metrics", "/health", "/live"]
    
    async def dispatch(self, request: Request, call_next):
        """
        Processa a requisição, registrando métricas de latência e contagem.
        
        Args:
            request: Requisição HTTP
            call_next: Handler para próximo middleware
        """
        # Verifica se o caminho deve ser excluído
        path = request.url.path
        if path in self.exclude_paths:
            return await call_next(request)
        
        # Extrai informações de contexto
        tenant = getattr(request.state, "tenant", None) or "default"
        region = getattr(request.state, "region", None) or "global"
        
        # Normaliza o path para evitar cardinalidade alta
        normalized_path = self._normalize_path(path)
        
        # Registra o início da requisição
        start_time = time.time()
        
        # Incrementa contador de requisições
        self.metrics.http_requests_total.labels(
            method=request.method,
            path=normalized_path,
            tenant=tenant,
            region=region
        ).inc()
        
        # Processa a requisição
        try:
            response = await call_next(request)
            
            # Registra métricas de resposta
            status_code = response.status_code
            duration = time.time() - start_time
            
            # Incrementa contador de respostas por status
            self.metrics.http_responses_total.labels(
                method=request.method,
                path=normalized_path,
                status=str(status_code),
                tenant=tenant,
                region=region
            ).inc()
            
            # Registra duração da requisição
            self.metrics.http_request_duration_seconds.labels(
                method=request.method,
                path=normalized_path,
                tenant=tenant,
                region=region
            ).observe(duration)
            
            return response
            
        except Exception as exc:
            # Registra métricas de erro
            duration = time.time() - start_time
            
            self.metrics.http_exceptions_total.labels(
                method=request.method,
                path=normalized_path,
                exception=type(exc).__name__,
                tenant=tenant,
                region=region
            ).inc()
            
            self.metrics.http_request_duration_seconds.labels(
                method=request.method,
                path=normalized_path,
                tenant=tenant,
                region=region
            ).observe(duration)
            
            raise
    
    def _normalize_path(self, path: str) -> str:
        """
        Normaliza um path para evitar cardinalidade alta em métricas.
        
        Args:
            path: Path da URL
            
        Returns:
            Path normalizado
        """
        # Implementação básica - em produção, usar regex para substituir IDs por placeholders
        segments = path.strip('/').split('/')
        normalized = []
        
        for segment in segments:
            # Se o segmento parece um UUID, ID numérico, etc., substitui por placeholder
            if segment.isdigit() or (len(segment) > 8 and '-' in segment):
                normalized.append(":id")
            else:
                normalized.append(segment)
        
        normalized_path = '/' + '/'.join(normalized)
        return normalized_path if normalized_path != '/' else path


class ContextMiddleware(BaseHTTPMiddleware):
    """
    Middleware para propagação de contexto multi-tenant e multi-regional.
    
    Extrai informações de contexto de headers HTTP e os disponibiliza para
    os handlers e outros middlewares.
    """
    
    def __init__(
        self,
        app: FastAPI,
        default_tenant: str = "default",
        default_region: str = "global"
    ):
        """
        Inicializa o middleware de contexto.
        
        Args:
            app: Aplicação FastAPI
            default_tenant: Tenant padrão quando não especificado no header
            default_region: Região padrão quando não especificada no header
        """
        super().__init__(app)
        self.default_tenant = default_tenant
        self.default_region = default_region
    
    async def dispatch(self, request: Request, call_next):
        """
        Processa a requisição, extraindo e propagando informações de contexto.
        
        Args:
            request: Requisição HTTP
            call_next: Handler para próximo middleware
        """
        # Extrai informações de contexto dos headers
        tenant = request.headers.get("X-Tenant-ID", self.default_tenant)
        region = request.headers.get("X-Region", self.default_region)
        environment = request.headers.get("X-Environment", "production")
        
        # Armazena contexto no state da requisição
        request.state.tenant = tenant
        request.state.region = region
        request.state.environment = environment
        
        # Cria objeto de contexto para facilitar acesso
        request.state.context = ContextInfo(
            tenant=tenant,
            region=region,
            environment=environment
        )
        
        # Processa a requisição normalmente
        response = await call_next(request)
        
        # Adiciona headers de contexto na resposta para debugging
        response.headers["X-Tenant-ID"] = tenant
        response.headers["X-Region"] = region
        
        return response


def configure_observability(app: FastAPI, **kwargs) -> ObservabilityIntegration:
    """
    Configuração rápida de observabilidade para uma aplicação FastAPI.
    
    Args:
        app: Aplicação FastAPI a ser instrumentada
        **kwargs: Parâmetros adicionais para configuração
        
    Returns:
        Integração de observabilidade configurada
    """
    # Cria configuração
    config = ObservabilityConfig(**kwargs)
    
    # Inicializa integração
    obs = ObservabilityIntegration(config=config)
    
    # Instrumenta a aplicação
    obs.instrument_app(app)
    
    return obs


# Singleton global para acesso em outras partes do código
_integration_instance = None

def get_integration() -> ObservabilityIntegration:
    """
    Obtém a instância singleton da integração de observabilidade.
    
    Returns:
        Instância do ObservabilityIntegration
    """
    global _integration_instance
    
    if _integration_instance is None:
        _integration_instance = ObservabilityIntegration()
    
    return _integration_instance


def set_integration(integration: ObservabilityIntegration) -> None:
    """
    Define a instância singleton da integração de observabilidade.
    
    Args:
        integration: Instância do ObservabilityIntegration
    """
    global _integration_instance
    _integration_instance = integration