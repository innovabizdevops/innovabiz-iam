"""
Testes unitários para o módulo de health checks do framework de observabilidade.

Estes testes verificam o funcionamento dos health checks, verificações de dependências,
respostas de diagnóstico e suporte a contexto multi-tenant e multi-região.
"""

import unittest
from unittest import mock
import asyncio
from datetime import datetime, timedelta

import pytest
from fastapi.testclient import TestClient
from fastapi import FastAPI, Depends
from pydantic import BaseModel

from src.observability.config import HealthConfig
from src.observability.health import (
    HealthChecker,
    HealthStatus,
    ComponentType,
    ComponentStatus,
    ComponentCheckResult,
    HealthResponse,
    DiagnosticResponse
)
from src.observability.metrics import MetricsManager


class TestComponentCheckResult(unittest.TestCase):
    """Testes para a classe ComponentCheckResult."""

    def test_component_check_result_creation(self):
        """Testa a criação de resultados de verificação de componentes."""
        # Resultado saudável
        result = ComponentCheckResult(
            status=HealthStatus.HEALTHY,
            description="Database connection is healthy",
            latency_ms=42.5
        )
        self.assertEqual(result.status, HealthStatus.HEALTHY)
        self.assertEqual(result.description, "Database connection is healthy")
        self.assertEqual(result.latency_ms, 42.5)
        
        # Resultado não saudável com detalhes de erro
        result = ComponentCheckResult(
            status=HealthStatus.UNHEALTHY,
            description="Failed to connect to database",
            error="Connection timeout",
            latency_ms=500.2
        )
        self.assertEqual(result.status, HealthStatus.UNHEALTHY)
        self.assertEqual(result.description, "Failed to connect to database")
        self.assertEqual(result.error, "Connection timeout")
        self.assertEqual(result.latency_ms, 500.2)
    
    def test_component_check_result_dict_conversion(self):
        """Testa a conversão de resultados para dicionário."""
        result = ComponentCheckResult(
            status=HealthStatus.DEGRADED,
            description="Slow database response",
            latency_ms=350.0,
            details={"connection_pool": "exhausted"}
        )
        
        result_dict = result.dict()
        self.assertEqual(result_dict["status"], HealthStatus.DEGRADED)
        self.assertEqual(result_dict["description"], "Slow database response")
        self.assertEqual(result_dict["latency_ms"], 350.0)
        self.assertEqual(result_dict["details"]["connection_pool"], "exhausted")


@pytest.mark.asyncio
class TestHealthChecker:
    """Testes para o verificador de saúde."""

    @pytest.fixture
    async def health_checker(self):
        """Configura um verificador de saúde para testes."""
        # Mock do gerenciador de métricas
        metrics_manager = mock.MagicMock(spec=MetricsManager)
        
        # Configuração para health checks
        config = HealthConfig(
            enabled=True,
            cache_time_seconds=5,
            timeout_seconds=1.0,
            include_details_in_health=True
        )
        
        # Cria o health checker
        health_checker = HealthChecker(
            config=config,
            metrics=metrics_manager,
            service_start_time=datetime.utcnow()
        )
        
        yield health_checker
    
    async def test_register_dependency_checker(self, health_checker):
        """Testa o registro de verificadores de dependências."""
        # Define um verificador simples
        async def check_test_service(tenant, region):
            await asyncio.sleep(0.01)
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description="Test service is working",
                latency_ms=10.0
            )
        
        # Registra o verificador
        health_checker.register_dependency_checker(
            "test_service",
            check_test_service,
            {
                "name": "Test Service",
                "type": ComponentType.SERVICE,
                "critical": True
            }
        )
        
        # Verifica se o verificador foi registrado
        assert "test_service" in health_checker.dependency_checkers
        assert health_checker.dependency_metadata["test_service"]["name"] == "Test Service"
        assert health_checker.dependency_metadata["test_service"]["type"] == ComponentType.SERVICE
        assert health_checker.dependency_metadata["test_service"]["critical"] is True
    
    async def test_check_health(self, health_checker):
        """Testa a verificação básica de saúde."""
        # Define alguns verificadores com diferentes estados
        async def check_healthy(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description="Healthy component",
                latency_ms=5.0
            )
            
        async def check_degraded(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.DEGRADED,
                description="Degraded component",
                latency_ms=250.0
            )
            
        async def check_unhealthy(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.UNHEALTHY,
                description="Unhealthy component",
                error="Connection refused",
                latency_ms=500.0
            )
        
        # Registra os verificadores
        health_checker.register_dependency_checker(
            "healthy_component", 
            check_healthy,
            {"name": "Healthy Component", "type": ComponentType.SERVICE, "critical": False}
        )
        
        health_checker.register_dependency_checker(
            "degraded_component", 
            check_degraded,
            {"name": "Degraded Component", "type": ComponentType.SERVICE, "critical": False}
        )
        
        health_checker.register_dependency_checker(
            "unhealthy_component", 
            check_unhealthy,
            {"name": "Unhealthy Component", "type": ComponentType.SERVICE, "critical": True}
        )
        
        # Executa a verificação de saúde
        health_response = await health_checker.check_health("test_tenant", "test_region")
        
        # Verifica se a resposta é do tipo correto
        assert isinstance(health_response, HealthResponse)
        
        # Como temos um componente crítico não saudável, o status geral deve ser UNHEALTHY
        assert health_response.status == HealthStatus.UNHEALTHY
        
        # Verifica se todos os componentes estão nos resultados
        component_keys = {key for key in health_response.components.keys()}
        assert component_keys == {"healthy_component", "degraded_component", "unhealthy_component"}
        
        # Verifica o status de cada componente
        assert health_response.components["healthy_component"].status == HealthStatus.HEALTHY
        assert health_response.components["degraded_component"].status == HealthStatus.DEGRADED
        assert health_response.components["unhealthy_component"].status == HealthStatus.UNHEALTHY
    
    async def test_check_readiness(self, health_checker):
        """Testa a verificação de readiness."""
        # Define verificadores para readiness
        async def check_database(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description="Database is ready",
                latency_ms=10.0
            )
            
        async def check_cache(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description="Cache is ready",
                latency_ms=5.0
            )
        
        # Registra os verificadores
        health_checker.register_dependency_checker(
            "database", 
            check_database,
            {"name": "Database", "type": ComponentType.DATABASE, "critical": True}
        )
        
        health_checker.register_dependency_checker(
            "cache", 
            check_cache,
            {"name": "Cache", "type": ComponentType.CACHE, "critical": False}
        )
        
        # Executa a verificação de readiness
        readiness_response = await health_checker.check_readiness("test_tenant", "test_region")
        
        # Verifica se a resposta é do tipo correto
        assert isinstance(readiness_response, HealthResponse)
        
        # Como todos os componentes estão saudáveis, o status geral deve ser HEALTHY
        assert readiness_response.status == HealthStatus.HEALTHY
        
        # Verifica se todos os componentes estão nos resultados
        component_keys = {key for key in readiness_response.components.keys()}
        assert component_keys == {"database", "cache"}
    
    async def test_result_caching(self, health_checker):
        """Testa se os resultados estão sendo cacheados corretamente."""
        # Contador para rastrear chamadas ao verificador
        call_count = 0
        
        # Define um verificador que incrementa o contador
        async def check_with_counter(tenant, region):
            nonlocal call_count
            call_count += 1
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description=f"Call count: {call_count}",
                latency_ms=5.0
            )
        
        # Registra o verificador com tempo de cache baixo para teste
        health_checker.config.cache_time_seconds = 0.1
        health_checker.register_dependency_checker(
            "counter_service", 
            check_with_counter,
            {"name": "Counter Service", "type": ComponentType.SERVICE}
        )
        
        # Primeira chamada
        response1 = await health_checker.check_health("test_tenant", "test_region")
        assert call_count == 1
        assert response1.components["counter_service"].description == "Call count: 1"
        
        # Segunda chamada imediata (deve usar cache)
        response2 = await health_checker.check_health("test_tenant", "test_region")
        assert call_count == 1  # Não deve ter incrementado
        assert response2.components["counter_service"].description == "Call count: 1"
        
        # Espera o cache expirar
        await asyncio.sleep(0.2)
        
        # Terceira chamada (deve chamar o verificador novamente)
        response3 = await health_checker.check_health("test_tenant", "test_region")
        assert call_count == 2  # Deve ter incrementado
        assert response3.components["counter_service"].description == "Call count: 2"
    
    async def test_different_tenant_regions(self, health_checker):
        """Testa se o cache é isolado por tenant e região."""
        # Contador para rastrear chamadas ao verificador
        tenant_region_calls = {}
        
        # Define um verificador que registra chamadas por tenant e região
        async def check_tenant_region(tenant, region):
            key = f"{tenant}:{region}"
            tenant_region_calls[key] = tenant_region_calls.get(key, 0) + 1
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description=f"Tenant: {tenant}, Region: {region}, Calls: {tenant_region_calls[key]}",
                latency_ms=5.0
            )
        
        # Registra o verificador
        health_checker.register_dependency_checker(
            "tenant_region_service", 
            check_tenant_region,
            {"name": "Multi-tenant Service", "type": ComponentType.SERVICE}
        )
        
        # Chamadas para diferentes combinações de tenant e região
        await health_checker.check_health("tenant1", "region1")
        await health_checker.check_health("tenant1", "region2")
        await health_checker.check_health("tenant2", "region1")
        
        # Verifica se cada combinação foi chamada uma vez
        assert tenant_region_calls["tenant1:region1"] == 1
        assert tenant_region_calls["tenant1:region2"] == 1
        assert tenant_region_calls["tenant2:region1"] == 1
        
        # Chama novamente com a mesma combinação (deve usar cache)
        await health_checker.check_health("tenant1", "region1")
        
        # Não deve ter chamado novamente para tenant1:region1
        assert tenant_region_calls["tenant1:region1"] == 1
    
    async def test_get_diagnostic(self, health_checker):
        """Testa a geração de relatório de diagnóstico detalhado."""
        # Define alguns verificadores simples
        async def check_database(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description="Database is working",
                latency_ms=10.0,
                details={"connection_pool": "active"}
            )
            
        async def check_service(tenant, region):
            return ComponentCheckResult(
                status=HealthStatus.DEGRADED,
                description="External service is slow",
                latency_ms=300.0,
                details={"response_time": "degraded"}
            )
        
        # Registra os verificadores
        health_checker.register_dependency_checker(
            "database", 
            check_database,
            {"name": "Database", "type": ComponentType.DATABASE, "critical": True}
        )
        
        health_checker.register_dependency_checker(
            "external_service", 
            check_service,
            {"name": "External API", "type": ComponentType.SERVICE, "critical": False}
        )
        
        # Define o start time para uptime
        health_checker.service_start_time = datetime.utcnow() - timedelta(hours=2, minutes=30)
        
        # Gera o diagnóstico
        diagnostic = await health_checker.get_diagnostic(
            tenant="test_tenant",
            region="test_region",
            include_deps=True,
            include_metrics=True,
            include_config=True
        )
        
        # Verifica se a resposta é do tipo correto
        assert isinstance(diagnostic, DiagnosticResponse)
        
        # Verifica o status geral e uptime
        assert diagnostic.status == HealthStatus.DEGRADED  # Degradado devido ao serviço externo
        assert diagnostic.uptime_seconds > 9000  # 2h30m = 9000s
        
        # Verifica se os componentes estão incluídos
        assert "database" in diagnostic.components
        assert "external_service" in diagnostic.components
        
        # Verifica se os detalhes estão incluídos
        assert diagnostic.components["database"].details["connection_pool"] == "active"
        assert diagnostic.components["external_service"].details["response_time"] == "degraded"
        
        # Verifica se a configuração está incluída
        assert diagnostic.config is not None
    
    async def test_timeout_handling(self, health_checker):
        """Testa o tratamento de timeout em verificações."""
        # Define um verificador que demora mais que o timeout
        async def slow_checker(tenant, region):
            await asyncio.sleep(10.0)  # Demora muito tempo
            return ComponentCheckResult(
                status=HealthStatus.HEALTHY,
                description="This should timeout",
                latency_ms=10000.0
            )
        
        # Registra o verificador com timeout baixo
        health_checker.config.timeout_seconds = 0.1
        health_checker.register_dependency_checker(
            "slow_service", 
            slow_checker,
            {"name": "Slow Service", "type": ComponentType.SERVICE, "critical": True}
        )
        
        # Executa a verificação
        response = await health_checker.check_health("test_tenant", "test_region")
        
        # Verifica se o componente foi marcado como UNHEALTHY devido ao timeout
        assert response.components["slow_service"].status == HealthStatus.UNHEALTHY
        assert "timeout" in response.components["slow_service"].error.lower()


class TestFastAPIIntegration:
    """Testes para a integração com FastAPI."""
    
    @pytest.fixture
    def app_client(self):
        """Configura um cliente de teste FastAPI com health checks."""
        app = FastAPI()
        
        # Mock do health checker
        health_checker = mock.MagicMock(spec=HealthChecker)
        
        # Configura o comportamento dos mocks
        async def mock_check_health(tenant, region):
            return HealthResponse(
                status=HealthStatus.HEALTHY,
                timestamp=datetime.utcnow().isoformat(),
                components={
                    "database": ComponentCheckResult(
                        status=HealthStatus.HEALTHY,
                        description="Database is healthy",
                        latency_ms=5.0
                    )
                }
            )
            
        async def mock_check_readiness(tenant, region):
            return HealthResponse(
                status=HealthStatus.HEALTHY,
                timestamp=datetime.utcnow().isoformat(),
                components={
                    "database": ComponentCheckResult(
                        status=HealthStatus.HEALTHY,
                        description="Database is ready",
                        latency_ms=5.0
                    ),
                    "cache": ComponentCheckResult(
                        status=HealthStatus.HEALTHY,
                        description="Cache is ready",
                        latency_ms=3.0
                    )
                }
            )
            
        async def mock_get_diagnostic(tenant, region, include_deps, include_metrics, include_config):
            return DiagnosticResponse(
                status=HealthStatus.HEALTHY,
                timestamp=datetime.utcnow().isoformat(),
                uptime_seconds=3600,
                components={
                    "database": ComponentCheckResult(
                        status=HealthStatus.HEALTHY,
                        description="Database is healthy",
                        latency_ms=5.0
                    )
                },
                metrics={
                    "requests_total": 1000,
                    "errors_total": 10
                },
                config={
                    "timeout_seconds": 1.0
                }
            )
        
        # Atribui os mocks
        health_checker.check_health = mock_check_health
        health_checker.check_readiness = mock_check_readiness
        health_checker.get_diagnostic = mock_get_diagnostic
        
        # Define os endpoints
        @app.get("/health")
        async def health_endpoint(
            tenant: str = "default",
            region: str = "global"
        ):
            return await health_checker.check_health(tenant, region)
            
        @app.get("/ready")
        async def ready_endpoint(
            tenant: str = "default",
            region: str = "global"
        ):
            return await health_checker.check_readiness(tenant, region)
            
        @app.get("/diagnostic")
        async def diagnostic_endpoint(
            tenant: str = "default",
            region: str = "global",
            include_deps: bool = True,
            include_metrics: bool = True,
            include_config: bool = False
        ):
            return await health_checker.get_diagnostic(
                tenant=tenant,
                region=region,
                include_deps=include_deps,
                include_metrics=include_metrics,
                include_config=include_config
            )
            
        # Cria o cliente de teste
        client = TestClient(app)
        return client
    
    def test_health_endpoint(self, app_client):
        """Testa o endpoint de health check."""
        response = app_client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == HealthStatus.HEALTHY
        assert "components" in data
        assert "database" in data["components"]
    
    def test_ready_endpoint(self, app_client):
        """Testa o endpoint de readiness check."""
        response = app_client.get("/ready")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == HealthStatus.HEALTHY
        assert "components" in data
        assert "database" in data["components"]
        assert "cache" in data["components"]
    
    def test_diagnostic_endpoint(self, app_client):
        """Testa o endpoint de diagnóstico."""
        response = app_client.get("/diagnostic")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == HealthStatus.HEALTHY
        assert "uptime_seconds" in data
        assert data["uptime_seconds"] == 3600
        assert "components" in data
        assert "metrics" in data
        assert "config" in data
    
    def test_tenant_region_headers(self, app_client):
        """Testa a passagem de tenant e região via headers."""
        response = app_client.get(
            "/health",
            headers={"X-Tenant-ID": "custom_tenant", "X-Region": "custom_region"}
        )
        assert response.status_code == 200


if __name__ == "__main__":
    unittest.main()