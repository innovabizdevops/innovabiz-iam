"""
Módulo de health checks e diagnósticos para o IAM Audit Service.

Implementa verificações de saúde para componentes do sistema, incluindo
endpoints para verificações de liveness, readiness e health, além de
um endpoint de diagnóstico detalhado com métricas e estatísticas.

Design baseado no ADR-006: Health Checks e Endpoints de Diagnóstico.
"""

import time
import logging
import asyncio
from enum import Enum
from typing import Dict, List, Optional, Any, Callable, Union
from datetime import datetime, timedelta

from pydantic import BaseModel, Field

from .config import HealthConfig

logger = logging.getLogger(__name__)


class HealthStatus(str, Enum):
    """Possíveis status de saúde para componentes do sistema."""
    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"
    UNKNOWN = "unknown"


class ComponentCheckResult(BaseModel):
    """Resultado da verificação de um componente específico."""
    
    status: HealthStatus = HealthStatus.UNKNOWN
    description: Optional[str] = None
    latency_ms: Optional[float] = None
    last_check: datetime = Field(default_factory=datetime.utcnow)
    details: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


class DependencyInfo(BaseModel):
    """Informações detalhadas sobre uma dependência externa."""
    
    name: str
    type: str
    status: HealthStatus = HealthStatus.UNKNOWN
    latency_ms: Optional[float] = None
    last_success: Optional[datetime] = None
    last_failure: Optional[datetime] = None
    consecutive_failures: int = 0
    total_checks: int = 0
    total_failures: int = 0
    error_rate: float = 0.0
    details: Optional[Dict[str, Any]] = None


class HealthResponse(BaseModel):
    """Modelo de resposta para endpoints de health check."""
    
    status: HealthStatus
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    version: Optional[str] = None
    components: Dict[str, ComponentCheckResult] = {}
    
    # Campos multi-contexto
    tenant: Optional[str] = None
    region: Optional[str] = None
    environment: Optional[str] = "production"


class DiagnosticResponse(BaseModel):
    """Modelo de resposta para endpoint de diagnóstico detalhado."""
    
    status: HealthStatus
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    version: Optional[str] = None
    uptime_seconds: float = 0
    
    # Status detalhado de componentes
    components: Dict[str, ComponentCheckResult] = {}
    
    # Informações de dependências
    dependencies: Dict[str, DependencyInfo] = {}
    
    # Estatísticas e métricas
    metrics: Optional[Dict[str, Any]] = None
    
    # Informações de configuração (apenas se explicitamente solicitado)
    config: Optional[Dict[str, Any]] = None
    
    # Campos multi-contexto
    tenant: Optional[str] = None
    region: Optional[str] = None
    environment: Optional[str] = "production"


class HealthChecker:
    """
    Implementa verificações de saúde para componentes do sistema.
    
    Fornece endpoints para verificações de liveness, readiness e health,
    além de um endpoint de diagnóstico detalhado com métricas e estatísticas.
    """
    
    def __init__(
        self,
        config: Optional[HealthConfig] = None,
        metrics: Optional[Any] = None,
        service_start_time: Optional[datetime] = None
    ):
        """
        Inicializa o verificador de saúde.
        
        Args:
            config: Configuração para health checks
            metrics: Gerenciador de métricas (opcional)
            service_start_time: Horário de inicialização do serviço (opcional)
        """
        self.config = config or HealthConfig()
        self.metrics = metrics
        self.service_start_time = service_start_time or datetime.utcnow()
        
        # Registra verificadores para cada tipo de dependência
        self.dependency_checkers: Dict[str, Callable] = {}
        self.dependencies: Dict[str, DependencyInfo] = {}
        
        # Resultados em cache para evitar sobrecarga de verificações
        self._last_results: Dict[str, ComponentCheckResult] = {}
        self._cache_ttl = timedelta(seconds=5)  # Cache por 5 segundos
        self._lock = asyncio.Lock()  # Lock para evitar verificações concorrentes
    
    async def initialize(self):
        """Inicializa recursos e realiza verificações iniciais."""
        # Registra verificadores padrão
        await self._register_default_checkers()
        
        # Realiza verificação inicial
        try:
            await self.check_all_dependencies("default", "global")
        except Exception as e:
            logger.warning(f"Falha na verificação inicial de dependências: {e}")
    
    async def shutdown(self):
        """Limpa recursos ao encerrar."""
        # Nada a fazer por enquanto
        pass
    
    async def _register_default_checkers(self):
        """Registra verificadores padrão para dependências comuns."""
        # Registra verificador para o banco de dados
        self.register_dependency_checker(
            "database", 
            self._check_database_health,
            {
                "name": "PostgreSQL Database",
                "type": "database",
                "status": HealthStatus.UNKNOWN
            }
        )
        
        # Registra verificador para cache
        self.register_dependency_checker(
            "cache", 
            self._check_cache_health,
            {
                "name": "Redis Cache",
                "type": "cache",
                "status": HealthStatus.UNKNOWN
            }
        )
        
        # Registra verificador para Kafka
        self.register_dependency_checker(
            "kafka", 
            self._check_kafka_health,
            {
                "name": "Kafka Message Broker",
                "type": "message_broker",
                "status": HealthStatus.UNKNOWN
            }
        )
        
        # Registra verificador para storage
        self.register_dependency_checker(
            "storage", 
            self._check_storage_health,
            {
                "name": "Object Storage",
                "type": "storage",
                "status": HealthStatus.UNKNOWN
            }
        )
    
    def register_dependency_checker(
        self,
        dependency_name: str,
        checker_func: Callable,
        dependency_info: Dict[str, Any] = None
    ):
        """
        Registra um verificador para uma dependência externa.
        
        Args:
            dependency_name: Nome da dependência
            checker_func: Função assíncrona para verificar a saúde
            dependency_info: Informações iniciais sobre a dependência
        """
        self.dependency_checkers[dependency_name] = checker_func
        
        # Inicializa informações da dependência
        if dependency_name not in self.dependencies:
            if dependency_info:
                self.dependencies[dependency_name] = DependencyInfo(**dependency_info)
            else:
                self.dependencies[dependency_name] = DependencyInfo(
                    name=dependency_name,
                    type="unknown",
                    status=HealthStatus.UNKNOWN
                )
    
    async def _check_database_health(self, tenant: str, region: str) -> ComponentCheckResult:
        """
        Verifica a saúde do banco de dados.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Resultado da verificação
        """
        # TODO: Implementar verificação real com um cliente de banco de dados
        # Simulação para desenvolvimento
        start_time = time.time()
        await asyncio.sleep(0.01)  # Simulação de latência
        
        # Em uma implementação real:
        # - Verificar conexão
        # - Executar uma consulta simples
        # - Verificar latência
        
        latency_ms = (time.time() - start_time) * 1000
        
        return ComponentCheckResult(
            status=HealthStatus.HEALTHY,
            description="Database connection is working",
            latency_ms=latency_ms,
            last_check=datetime.utcnow(),
            details={
                "tenant": tenant,
                "region": region,
                "connection_pool": {
                    "available": 5,
                    "used": 2,
                    "max": 10
                }
            }
        )
    
    async def _check_cache_health(self, tenant: str, region: str) -> ComponentCheckResult:
        """
        Verifica a saúde do sistema de cache.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Resultado da verificação
        """
        # TODO: Implementar verificação real com um cliente de cache
        # Simulação para desenvolvimento
        start_time = time.time()
        await asyncio.sleep(0.005)  # Simulação de latência
        
        latency_ms = (time.time() - start_time) * 1000
        
        return ComponentCheckResult(
            status=HealthStatus.HEALTHY,
            description="Cache connection is working",
            latency_ms=latency_ms,
            last_check=datetime.utcnow(),
            details={
                "tenant": tenant,
                "region": region,
                "hit_rate": 0.92,
                "memory_usage_mb": 128
            }
        )
    
    async def _check_kafka_health(self, tenant: str, region: str) -> ComponentCheckResult:
        """
        Verifica a saúde do broker Kafka.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Resultado da verificação
        """
        # TODO: Implementar verificação real com um cliente Kafka
        # Simulação para desenvolvimento
        start_time = time.time()
        await asyncio.sleep(0.008)  # Simulação de latência
        
        latency_ms = (time.time() - start_time) * 1000
        
        return ComponentCheckResult(
            status=HealthStatus.HEALTHY,
            description="Kafka connection is working",
            latency_ms=latency_ms,
            last_check=datetime.utcnow(),
            details={
                "tenant": tenant,
                "region": region,
                "topics": {
                    "audit_events": {"lag": 0, "partitions": 3},
                    "retention_events": {"lag": 2, "partitions": 3}
                }
            }
        )
    
    async def _check_storage_health(self, tenant: str, region: str) -> ComponentCheckResult:
        """
        Verifica a saúde do sistema de armazenamento.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Resultado da verificação
        """
        # TODO: Implementar verificação real com um cliente de storage
        # Simulação para desenvolvimento
        start_time = time.time()
        await asyncio.sleep(0.015)  # Simulação de latência
        
        latency_ms = (time.time() - start_time) * 1000
        
        return ComponentCheckResult(
            status=HealthStatus.HEALTHY,
            description="Storage connection is working",
            latency_ms=latency_ms,
            last_check=datetime.utcnow(),
            details={
                "tenant": tenant,
                "region": region,
                "usage_gb": 1250,
                "available_gb": 8750
            }
        )
    
    async def check_component(
        self,
        component: str,
        tenant: str,
        region: str,
        force: bool = False
    ) -> ComponentCheckResult:
        """
        Verifica a saúde de um componente específico.
        
        Args:
            component: Nome do componente a ser verificado
            tenant: ID do tenant
            region: Código da região
            force: Força nova verificação mesmo se houver cache recente
            
        Returns:
            Resultado da verificação
        """
        cache_key = f"{component}:{tenant}:{region}"
        
        # Verifica cache, a menos que force=True
        if not force and cache_key in self._last_results:
            last_check = self._last_results[cache_key]
            if datetime.utcnow() - last_check.last_check < self._cache_ttl:
                return last_check
        
        # Verifica se existe verificador para o componente
        if component not in self.dependency_checkers:
            result = ComponentCheckResult(
                status=HealthStatus.UNKNOWN,
                description=f"No health checker for component: {component}",
                last_check=datetime.utcnow()
            )
            self._last_results[cache_key] = result
            return result
        
        # Realiza a verificação com timeout
        checker = self.dependency_checkers[component]
        start_time = time.time()
        
        try:
            # Define timeout baseado na configuração
            timeout_ms = getattr(self.config, f"{component}_timeout_ms", 1000)
            timeout_sec = timeout_ms / 1000.0
            
            # Executa a verificação com timeout
            result = await asyncio.wait_for(
                checker(tenant, region),
                timeout=timeout_sec
            )
            
            # Atualiza informações da dependência
            if component in self.dependencies:
                dep = self.dependencies[component]
                dep.status = result.status
                dep.latency_ms = result.latency_ms
                dep.last_success = datetime.utcnow()
                dep.total_checks += 1
                
                # Calcula taxa de erro
                if dep.total_checks > 0:
                    dep.error_rate = dep.total_failures / dep.total_checks
                
                # Reseta contagem de falhas consecutivas
                dep.consecutive_failures = 0
            
            # Atualiza cache
            self._last_results[cache_key] = result
            
            # Incrementa métricas
            if self.metrics:
                self.metrics.health_check_total.labels(
                    check_type=component,
                    status="success",
                    tenant=tenant,
                    region=region
                ).inc()
            
            return result
            
        except asyncio.TimeoutError:
            # Cria resultado para timeout
            latency_ms = (time.time() - start_time) * 1000
            result = ComponentCheckResult(
                status=HealthStatus.DEGRADED,
                description=f"Health check timed out after {latency_ms:.2f}ms",
                latency_ms=latency_ms,
                last_check=datetime.utcnow(),
                error=f"Timeout after {latency_ms:.2f}ms"
            )
            
            # Atualiza informações da dependência
            if component in self.dependencies:
                dep = self.dependencies[component]
                dep.status = HealthStatus.DEGRADED
                dep.latency_ms = latency_ms
                dep.last_failure = datetime.utcnow()
                dep.consecutive_failures += 1
                dep.total_failures += 1
                dep.total_checks += 1
                
                # Calcula taxa de erro
                if dep.total_checks > 0:
                    dep.error_rate = dep.total_failures / dep.total_checks
            
            # Atualiza cache
            self._last_results[cache_key] = result
            
            # Incrementa métricas
            if self.metrics:
                self.metrics.health_check_total.labels(
                    check_type=component,
                    status="timeout",
                    tenant=tenant,
                    region=region
                ).inc()
                
                self.metrics.dependency_check_failed.labels(
                    dependency=component,
                    tenant=tenant,
                    region=region
                ).inc()
            
            return result
            
        except Exception as e:
            # Cria resultado para erro
            latency_ms = (time.time() - start_time) * 1000
            result = ComponentCheckResult(
                status=HealthStatus.UNHEALTHY,
                description=f"Health check failed: {str(e)}",
                latency_ms=latency_ms,
                last_check=datetime.utcnow(),
                error=str(e)
            )
            
            # Atualiza informações da dependência
            if component in self.dependencies:
                dep = self.dependencies[component]
                dep.status = HealthStatus.UNHEALTHY
                dep.latency_ms = latency_ms
                dep.last_failure = datetime.utcnow()
                dep.consecutive_failures += 1
                dep.total_failures += 1
                dep.total_checks += 1
                
                # Calcula taxa de erro
                if dep.total_checks > 0:
                    dep.error_rate = dep.total_failures / dep.total_checks
            
            # Atualiza cache
            self._last_results[cache_key] = result
            
            # Incrementa métricas
            if self.metrics:
                self.metrics.health_check_total.labels(
                    check_type=component,
                    status="error",
                    tenant=tenant,
                    region=region
                ).inc()
                
                self.metrics.dependency_check_failed.labels(
                    dependency=component,
                    tenant=tenant,
                    region=region
                ).inc()
            
            # Registra o erro
            logger.error(f"Health check failed for {component}: {str(e)}")
            
            return result
    
    async def check_health(self, tenant: str, region: str) -> HealthResponse:
        """
        Realiza verificações básicas de saúde para componentes essenciais.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Resposta de health check
        """
        results = {}
        overall_status = HealthStatus.HEALTHY
        
        # Verifica apenas componentes configurados para health check
        components = self.config.health_components
        
        # Executa verificações em paralelo
        tasks = [
            self.check_component(component, tenant, region)
            for component in components
        ]
        
        if tasks:
            component_results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Processa resultados
            for i, result in enumerate(component_results):
                component = components[i]
                
                if isinstance(result, Exception):
                    # Trata exceção na verificação
                    results[component] = ComponentCheckResult(
                        status=HealthStatus.UNHEALTHY,
                        description=f"Check failed: {str(result)}",
                        last_check=datetime.utcnow(),
                        error=str(result)
                    )
                    overall_status = HealthStatus.DEGRADED
                else:
                    # Armazena resultado normal
                    results[component] = result
                    
                    # Atualiza status geral
                    if result.status == HealthStatus.UNHEALTHY:
                        overall_status = HealthStatus.UNHEALTHY
                    elif result.status == HealthStatus.DEGRADED and overall_status != HealthStatus.UNHEALTHY:
                        overall_status = HealthStatus.DEGRADED
        
        # Cria resposta
        return HealthResponse(
            status=overall_status,
            timestamp=datetime.utcnow(),
            components=results,
            tenant=tenant,
            region=region
        )
    
    async def check_readiness(self, tenant: str, region: str) -> HealthResponse:
        """
        Verifica se o serviço está pronto para receber requisições.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Resposta de readiness check
        """
        results = {}
        overall_status = HealthStatus.HEALTHY
        
        # Verifica apenas componentes configurados para readiness check
        components = self.config.readiness_components
        
        # Executa verificações em paralelo
        tasks = [
            self.check_component(component, tenant, region)
            for component in components
        ]
        
        if tasks:
            component_results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Processa resultados
            for i, result in enumerate(component_results):
                component = components[i]
                
                if isinstance(result, Exception):
                    # Trata exceção na verificação
                    results[component] = ComponentCheckResult(
                        status=HealthStatus.UNHEALTHY,
                        description=f"Check failed: {str(result)}",
                        last_check=datetime.utcnow(),
                        error=str(result)
                    )
                    overall_status = HealthStatus.UNHEALTHY  # Falha de readiness é mais grave
                else:
                    # Armazena resultado normal
                    results[component] = result
                    
                    # Atualiza status geral - para readiness, qualquer componente não saudável torna o serviço não pronto
                    if result.status != HealthStatus.HEALTHY:
                        overall_status = HealthStatus.UNHEALTHY
        
        # Cria resposta
        return HealthResponse(
            status=overall_status,
            timestamp=datetime.utcnow(),
            components=results,
            tenant=tenant,
            region=region
        )
    
    async def check_liveness(self) -> HealthResponse:
        """
        Verifica se o processo está vivo e responsivo.
        
        Returns:
            Resposta de liveness check
        """
        # Liveness é uma verificação simples apenas para confirmar que o processo está respondendo
        return HealthResponse(
            status=HealthStatus.HEALTHY,
            timestamp=datetime.utcnow(),
            components={}
        )
    
    async def check_all_dependencies(self, tenant: str, region: str) -> Dict[str, ComponentCheckResult]:
        """
        Verifica a saúde de todas as dependências registradas.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            
        Returns:
            Mapa de resultados das verificações
        """
        results = {}
        
        # Executa verificações em paralelo
        tasks = [
            self.check_component(component, tenant, region, force=True)
            for component in self.dependency_checkers.keys()
        ]
        
        if tasks:
            component_results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Processa resultados
            for i, result in enumerate(component_results):
                component = list(self.dependency_checkers.keys())[i]
                
                if isinstance(result, Exception):
                    # Trata exceção na verificação
                    results[component] = ComponentCheckResult(
                        status=HealthStatus.UNHEALTHY,
                        description=f"Check failed: {str(result)}",
                        last_check=datetime.utcnow(),
                        error=str(result)
                    )
                else:
                    # Armazena resultado normal
                    results[component] = result
        
        return results
    
    async def get_diagnostic(
        self,
        tenant: str,
        region: str,
        include_deps: bool = True,
        include_metrics: bool = True,
        include_config: bool = False
    ) -> DiagnosticResponse:
        """
        Obtém diagnóstico detalhado do serviço.
        
        Args:
            tenant: ID do tenant
            region: Código da região
            include_deps: Se deve incluir informações detalhadas de dependências
            include_metrics: Se deve incluir métricas
            include_config: Se deve incluir informações de configuração
            
        Returns:
            Resposta de diagnóstico
        """
        # Calcula tempo de atividade
        uptime_seconds = (datetime.utcnow() - self.service_start_time).total_seconds()
        
        # Obtém resultados das verificações
        component_results = await self.check_all_dependencies(tenant, region)
        
        # Determina status geral
        overall_status = HealthStatus.HEALTHY
        for result in component_results.values():
            if result.status == HealthStatus.UNHEALTHY:
                overall_status = HealthStatus.UNHEALTHY
                break
            elif result.status == HealthStatus.DEGRADED and overall_status != HealthStatus.UNHEALTHY:
                overall_status = HealthStatus.DEGRADED
        
        # Prepara resposta
        response = DiagnosticResponse(
            status=overall_status,
            timestamp=datetime.utcnow(),
            uptime_seconds=uptime_seconds,
            components=component_results,
            tenant=tenant,
            region=region
        )
        
        # Inclui dependências se solicitado
        if include_deps:
            response.dependencies = self.dependencies
        
        # Inclui métricas se solicitado
        if include_metrics and self.metrics:
            # TODO: Implementar coleta de métricas básicas
            response.metrics = {
                # Exemplo de métricas que seriam coletadas
                "http_requests_total": 12345,
                "http_errors_total": 123,
                "audit_events_processed": 5678,
                "average_response_time_ms": 42.5
            }
        
        # Inclui configuração se solicitado
        if include_config:
            # Filtra informações sensíveis
            response.config = {
                "metrics_enabled": self.config.enabled,
                "health_components": self.config.health_components,
                "readiness_components": self.config.readiness_components,
                "timeouts": {
                    "database": self.config.database_timeout_ms,
                    "cache": self.config.cache_timeout_ms,
                    "kafka": self.config.kafka_timeout_ms,
                    "storage": self.config.storage_timeout_ms
                }
            }
        
        return response