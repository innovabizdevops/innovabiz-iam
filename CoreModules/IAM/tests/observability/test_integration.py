"""
Testes unitários para a integração principal do framework de observabilidade.

Estes testes verificam a classe central ObservabilityIntegration e os middlewares
associados, garantindo a integração completa de métricas, health checks e tracing
em aplicações FastAPI.
"""

import unittest
from unittest import mock
import asyncio
from datetime import datetime

import pytest
from fastapi import FastAPI, Request, Response
from fastapi.testclient import TestClient
from prometheus_client import CollectorRegistry
from prometheus_client.parser import text_string_to_metric_families
from starlette.middleware.base import BaseHTTPMiddleware

from src.observability.config import (
    ObservabilityConfig,
    MetricsConfig,
    HealthConfig,
    TracingConfig
)
from src.observability.integration import (
    ObservabilityIntegration,
    HTTPMetricsMiddleware,
    ContextMiddleware
)


class TestContextMiddleware:
    """Testes para o middleware de contexto multi-tenant e multi-região."""

    @pytest.fixture
    def app_with_context_middleware(self):
        """Configura uma aplicação FastAPI com o middleware de contexto."""
        app = FastAPI()
        
        # Adiciona o middleware de contexto
        app.add_middleware(
            ContextMiddleware,
            default_tenant="default_tenant",
            default_region="default_region",
            default_environment="test"
        )
        
        # Adiciona uma rota que retorna os headers recebidos
        @app.get("/echo-headers")
        async def echo_headers(request: Request):
            return {
                "tenant": request.headers.get("X-Tenant-ID"),
                "region": request.headers.get("X-Region"),
                "environment": request.headers.get("X-Environment")
            }
        
        # Adiciona uma rota que cria uma resposta com contexto
        @app.get("/with-context")
        async def with_context(request: Request):
            return Response(
                content="OK",
                status_code=200,
                headers={
                    "X-Test": "test-value"
                }
            )
        
        return TestClient(app)
    
    def test_header_propagation(self, app_with_context_middleware):
        """Testa a propagação de headers de contexto."""
        # Faz uma requisição com headers personalizados
        response = app_with_context_middleware.get(
            "/echo-headers",
            headers={
                "X-Tenant-ID": "custom_tenant",
                "X-Region": "custom_region",
                "X-Environment": "custom_env"
            }
        )
        
        assert response.status_code == 200
        data = response.json()
        
        # Verifica se os headers foram recebidos corretamente
        assert data["tenant"] == "custom_tenant"
        assert data["region"] == "custom_region"
        assert data["environment"] == "custom_env"
    
    def test_default_values(self, app_with_context_middleware):
        """Testa os valores padrão quando headers não são fornecidos."""
        # Faz uma requisição sem headers personalizados
        response = app_with_context_middleware.get("/echo-headers")
        
        assert response.status_code == 200
        data = response.json()
        
        # Verifica se os valores padrão foram usados
        assert data["tenant"] == "default_tenant"
        assert data["region"] == "default_region"
        assert data["environment"] == "test"
    
    def test_response_headers(self, app_with_context_middleware):
        """Testa se os headers são adicionados à resposta."""
        # Faz uma requisição com headers personalizados
        response = app_with_context_middleware.get(
            "/with-context",
            headers={
                "X-Tenant-ID": "response_tenant",
                "X-Region": "response_region"
            }
        )
        
        assert response.status_code == 200
        
        # Verifica se os headers foram adicionados à resposta
        assert response.headers.get("X-Tenant-ID") == "response_tenant"
        assert response.headers.get("X-Region") == "response_region"
        
        # Verifica se o header original foi preservado
        assert response.headers.get("X-Test") == "test-value"


class TestHTTPMetricsMiddleware:
    """Testes para o middleware de métricas HTTP."""

    @pytest.fixture
    def registry_and_middleware(self):
        """Configura o registry e middleware para testes."""
        registry = CollectorRegistry()
        
        middleware = HTTPMetricsMiddleware(
            registry=registry,
            namespace="test",
            subsystem="http",
            default_labels={
                "service": "test-service",
                "version": "1.0.0"
            },
            exclude_paths=["/metrics", "/health"],
            path_normalization=True
        )
        
        return registry, middleware
    
    @pytest.fixture
    def app_with_metrics_middleware(self, registry_and_middleware):
        """Configura uma aplicação FastAPI com o middleware de métricas."""
        registry, middleware = registry_and_middleware
        app = FastAPI()
        
        # Adiciona o middleware diretamente
        app.add_middleware(BaseHTTPMiddleware, dispatch=middleware)
        
        # Adiciona uma rota de teste
        @app.get("/test")
        async def test_endpoint():
            return {"status": "ok"}
        
        # Adiciona uma rota com parâmetro para testar normalização
        @app.get("/users/{user_id}")
        async def get_user(user_id: str):
            return {"user_id": user_id}
        
        # Adiciona uma rota que causa erro
        @app.get("/error")
        async def error_endpoint():
            raise ValueError("Test error")
        
        # Adiciona uma rota excluída
        @app.get("/metrics")
        async def metrics_endpoint():
            return {"message": "This should not be instrumented"}
        
        # Rota para expor métricas
        @app.get("/export-metrics")
        def export_metrics():
            from prometheus_client import generate_latest
            return Response(content=generate_latest(registry).decode("utf-8"))
        
        return TestClient(app)
    
    def test_request_counting(self, app_with_metrics_middleware):
        """Testa a contagem de requisições."""
        # Faz algumas requisições
        app_with_metrics_middleware.get("/test")
        app_with_metrics_middleware.get("/test")
        app_with_metrics_middleware.get("/users/123")
        
        try:
            app_with_metrics_middleware.get("/error")
        except:
            pass
        
        # Pega as métricas geradas
        response = app_with_metrics_middleware.get("/export-metrics")
        metrics_text = response.text
        
        # Verifica se o contador de requisições foi incrementado
        for family in text_string_to_metric_families(metrics_text):
            if family.name == "test_http_requests_total":
                # Conta o total de amostras
                request_count = sum(sample.value for sample in family.samples)
                assert request_count == 4  # 3 sucessos + 1 erro
    
    def test_path_normalization(self, app_with_metrics_middleware):
        """Testa a normalização de caminhos."""
        # Faz algumas requisições para endpoints com parâmetros
        app_with_metrics_middleware.get("/users/123")
        app_with_metrics_middleware.get("/users/456")
        app_with_metrics_middleware.get("/users/789")
        
        # Pega as métricas geradas
        response = app_with_metrics_middleware.get("/export-metrics")
        metrics_text = response.text
        
        # Verifica se existe apenas uma métrica normalizada para /users/:id
        path_samples = []
        for family in text_string_to_metric_families(metrics_text):
            if family.name == "test_http_requests_total":
                for sample in family.samples:
                    if "path" in sample.labels:
                        path_samples.append(sample.labels["path"])
        
        # Deve ter apenas um caminho normalizado para todas as requisições /users/123, /users/456, etc.
        assert "/users/:id" in path_samples
        assert "/users/123" not in path_samples
        assert "/users/456" not in path_samples
    
    def test_exclude_paths(self, app_with_metrics_middleware):
        """Testa a exclusão de caminhos das métricas."""
        # Faz requisição para um caminho excluído
        app_with_metrics_middleware.get("/metrics")
        
        # Pega as métricas geradas
        response = app_with_metrics_middleware.get("/export-metrics")
        metrics_text = response.text
        
        # Verifica se o caminho /metrics não está nas métricas
        path_samples = []
        for family in text_string_to_metric_families(metrics_text):
            if family.name == "test_http_requests_total":
                for sample in family.samples:
                    if "path" in sample.labels:
                        path_samples.append(sample.labels["path"])
        
        assert "/metrics" not in path_samples
    
    def test_request_duration(self, app_with_metrics_middleware):
        """Testa o histograma de duração de requisições."""
        # Faz uma requisição
        app_with_metrics_middleware.get("/test")
        
        # Pega as métricas geradas
        response = app_with_metrics_middleware.get("/export-metrics")
        metrics_text = response.text
        
        # Verifica se o histograma de duração foi criado
        histogram_found = False
        for family in text_string_to_metric_families(metrics_text):
            if family.name == "test_http_request_duration_seconds":
                histogram_found = True
                break
        
        assert histogram_found


class TestObservabilityIntegration:
    """Testes para a classe central de integração de observabilidade."""

    @pytest.fixture
    def observability_config(self):
        """Configura a observabilidade para testes."""
        return ObservabilityConfig(
            service_name="test-service",
            metrics=MetricsConfig(
                enabled=True,
                namespace="test",
                subsystem="audit",
            ),
            health=HealthConfig(
                enabled=True,
                cache_time_seconds=5,
                timeout_seconds=1.0
            ),
            tracing=TracingConfig(
                enabled=True,
                service_name="test-service",
                sample_ratio=1.0
            )
        )
    
    @pytest.fixture
    def observability(self, observability_config):
        """Configura a integração de observabilidade para testes."""
        # Cria a integração com configuração de teste
        registry = CollectorRegistry()
        integration = ObservabilityIntegration(
            config=observability_config,
            registry=registry
        )
        
        # Mock para métodos assíncronos
        integration.initialize = mock.AsyncMock()
        integration.shutdown = mock.AsyncMock()
        
        return integration
    
    @pytest.fixture
    def app_with_observability(self, observability):
        """Configura uma aplicação FastAPI com observabilidade."""
        app = FastAPI()
        
        # Instrumenta o app com observabilidade
        observability.instrument_app(app)
        
        # Adiciona uma rota de teste
        @app.get("/test")
        async def test_endpoint():
            return {"status": "ok"}
        
        # Adiciona uma rota que usa contexto de tenant e região
        @app.get("/with-context")
        async def with_context(request: Request):
            tenant = request.headers.get("X-Tenant-ID", "default")
            region = request.headers.get("X-Region", "default")
            return {"tenant": tenant, "region": region}
        
        return TestClient(app)
    
    def test_middleware_registration(self, app_with_observability):
        """Testa se os middlewares são registrados corretamente."""
        # O app deve responder normalmente
        response = app_with_observability.get("/test")
        assert response.status_code == 200
        
        # Deve ter o middleware de contexto
        response = app_with_observability.get(
            "/with-context",
            headers={"X-Tenant-ID": "test_tenant", "X-Region": "test_region"}
        )
        assert response.status_code == 200
        data = response.json()
        assert data["tenant"] == "test_tenant"
        assert data["region"] == "test_region"
    
    def test_metrics_endpoint(self, app_with_observability):
        """Testa se o endpoint /metrics é configurado."""
        response = app_with_observability.get("/metrics")
        assert response.status_code == 200
        assert "TYPE" in response.text  # Padrão do formato Prometheus
    
    def test_health_endpoints(self, app_with_observability):
        """Testa se os endpoints de saúde são configurados."""
        # Endpoint /health
        response = app_with_observability.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        
        # Endpoint /ready
        response = app_with_observability.get("/ready")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        
        # Endpoint /live
        response = app_with_observability.get("/live")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
    
    def test_diagnostic_endpoint(self, app_with_observability):
        """Testa se o endpoint /diagnostic é configurado."""
        response = app_with_observability.get("/diagnostic")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        assert "uptime_seconds" in data
        
        # Verifica opções de query params
        response = app_with_observability.get(
            "/diagnostic?include_metrics=true&include_config=true"
        )
        assert response.status_code == 200
        data = response.json()
        assert "metrics" in data
        assert "config" in data
    
    def test_path_normalization(self, observability):
        """Testa a função de normalização de path."""
        # Testa vários cenários
        test_cases = [
            ("/users/123", "/users/:id"),
            ("/api/products/abc-123", "/api/products/:id"),
            ("/tenants/t1/resources/r2", "/tenants/:id/resources/:id"),
            ("/static/js/main.js", "/static/js/main.js"),
            ("/orders/order-123-xyz/items", "/orders/:id/items"),
            ("/events/2023-01-01/logs", "/events/:id/logs"),
            ("/audit/uuid-1234-5678-9012/details", "/audit/:id/details")
        ]
        
        for path, expected in test_cases:
            assert observability._normalize_path(path) == expected
    
    def test_register_dependencies(self, observability):
        """Testa o registro de dependências para health checks."""
        # Define funções de verificação simples
        async def check_database(tenant, region):
            return {"status": "healthy", "latency_ms": 10.0}
        
        async def check_cache(tenant, region):
            return {"status": "healthy", "latency_ms": 5.0}
        
        # Registra as dependências
        observability.register_health_dependency(
            "database",
            check_database,
            {"name": "PostgreSQL Database", "type": "database", "critical": True}
        )
        
        observability.register_health_dependency(
            "cache",
            check_cache,
            {"name": "Redis Cache", "type": "cache", "critical": False}
        )
        
        # Verifica se as dependências foram registradas
        assert "database" in observability.health_checker.dependency_checkers
        assert "cache" in observability.health_checker.dependency_checkers


class TestObservabilityLifecycle:
    """Testes para os eventos de ciclo de vida da observabilidade."""
    
    @pytest.fixture
    def observability_with_mocks(self):
        """Configura a integração com mocks para lifecycle."""
        config = ObservabilityConfig(
            service_name="test-service",
            metrics=MetricsConfig(enabled=True),
            health=HealthConfig(enabled=True),
            tracing=TracingConfig(enabled=True)
        )
        
        # Cria a integração
        integration = ObservabilityIntegration(config=config)
        
        # Adiciona mocks para os métodos de lifecycle
        integration.initialize = mock.AsyncMock()
        integration.shutdown = mock.AsyncMock()
        
        return integration
    
    @pytest.mark.asyncio
    async def test_lifecycle_handlers(self, observability_with_mocks):
        """Testa os handlers de ciclo de vida."""
        integration = observability_with_mocks
        
        # Simula o evento de startup
        await integration.startup_event_handler()
        integration.initialize.assert_called_once()
        
        # Simula o evento de shutdown
        await integration.shutdown_event_handler()
        integration.shutdown.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_initialize_shutdown(self):
        """Testa os métodos de inicialização e encerramento reais."""
        # Cria uma configuração com todos os componentes desabilitados
        # para evitar setup real durante os testes
        config = ObservabilityConfig(
            service_name="test-service",
            metrics=MetricsConfig(enabled=False),
            health=HealthConfig(enabled=False),
            tracing=TracingConfig(enabled=False)
        )
        
        # Cria a integração
        integration = ObservabilityIntegration(config=config)
        
        # Substitui os métodos internos por mocks
        integration._setup_metrics = mock.MagicMock()
        integration._setup_health = mock.MagicMock()
        integration._setup_tracing = mock.AsyncMock()
        integration._cleanup_tracing = mock.AsyncMock()
        
        # Testa inicialização
        await integration.initialize()
        integration._setup_metrics.assert_called_once()
        integration._setup_health.assert_called_once()
        integration._setup_tracing.assert_called_once()
        
        # Testa encerramento
        await integration.shutdown()
        integration._cleanup_tracing.assert_called_once()


if __name__ == "__main__":
    unittest.main()