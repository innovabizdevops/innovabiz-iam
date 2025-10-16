"""
Testes unitários para o módulo de integração de observabilidade do serviço IAM Audit.

Este módulo verifica o funcionamento correto da classe ObservabilityIntegration,
incluindo setup de métricas, health checks, manipuladores de eventos de ciclo de vida
e endpoints diagnósticos.

Author: INNOVABIZ DevOps Team
Date: 2025-07-31
"""

import json
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

import pytest
from fastapi import FastAPI, Response
from fastapi.testclient import TestClient
from prometheus_client import REGISTRY, Counter

from api.app.integrations.observability import ObservabilityIntegration, setup_observability
from api.app.metrics.audit_metrics import (
    AUDIT_SERVICE_INFO, HTTP_REQUESTS_TOTAL, HTTP_REQUEST_DURATION_SECONDS,
    AUDIT_EVENTS_PROCESSED_TOTAL
)


class TestObservabilityIntegration(unittest.TestCase):
    """Testes unitários para o módulo de integração de observabilidade."""

    def setUp(self):
        """Configuração para cada teste."""
        # Limpar registros de métricas do Prometheus entre testes
        for name in list(REGISTRY._names_to_collectors.keys()):
            if name.startswith('audit_') or name.startswith('http_'):
                REGISTRY.unregister(REGISTRY._names_to_collectors[name])

        # Criar aplicação FastAPI para testes
        self.app = FastAPI()
        self.client = TestClient(self.app)
        
        # Mock para dependências externas
        self.db_client_mock = AsyncMock()
        self.redis_client_mock = AsyncMock()
        self.kafka_client_mock = AsyncMock()
        self.storage_client_mock = AsyncMock()
        
        # Configurar todos os mocks para retornarem sucesso por padrão
        self.db_client_mock.is_healthy.return_value = True
        self.redis_client_mock.is_healthy.return_value = True
        self.kafka_client_mock.is_healthy.return_value = True
        self.storage_client_mock.is_healthy.return_value = True
        
        # Instância de observabilidade
        self.observability = ObservabilityIntegration(
            app=self.app,
            service_name="iam-audit-test",
            service_version="1.0.0",
            build_id="test-build",
            commit_hash="abcdef123456",
            environment="test",
            region="us-east",
            db_client=self.db_client_mock,
            redis_client=self.redis_client_mock,
            kafka_client=self.kafka_client_mock,
            storage_client=self.storage_client_mock,
        )

    def test_setup_metrics(self):
        """Verifica se as métricas são configuradas corretamente."""
        # Arrange & Act
        self.observability.setup_metrics()
        
        # Assert
        self.assertIsNotNone(AUDIT_SERVICE_INFO._metrics)
        self.assertIsNotNone(HTTP_REQUESTS_TOTAL._metrics)
        self.assertIsNotNone(HTTP_REQUEST_DURATION_SECONDS._metrics)
        self.assertIsNotNone(AUDIT_EVENTS_PROCESSED_TOTAL._metrics)

    def test_metrics_endpoint(self):
        """Verifica se o endpoint de métricas está funcionando."""
        # Arrange
        self.observability.setup_metrics()
        self.observability.setup_metrics_endpoint()
        
        # Act
        response = self.client.get("/metrics")
        
        # Assert
        self.assertEqual(response.status_code, 200)
        self.assertIn("application/openmetrics-text", response.headers["content-type"])
        self.assertIn("# HELP audit_service_info", response.text)

    @pytest.mark.asyncio
    async def test_health_check_all_healthy(self):
        """Verifica se o health check retorna OK quando todos os componentes estão saudáveis."""
        # Arrange
        self.observability.setup_health_check_endpoint()
        
        # Act
        response = self.client.get("/health")
        
        # Assert
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["status"], "healthy")
        self.assertEqual(data["service"], "iam-audit-test")
        self.assertEqual(len(data["checks"]), 4)  # DB, Redis, Kafka, Storage
        for check in data["checks"]:
            self.assertTrue(check["status"])

    @pytest.mark.asyncio
    async def test_health_check_db_unhealthy(self):
        """Verifica se o health check retorna erro quando o banco de dados está indisponível."""
        # Arrange
        self.db_client_mock.is_healthy.return_value = False
        self.observability.setup_health_check_endpoint()
        
        # Act
        response = self.client.get("/health")
        
        # Assert
        self.assertEqual(response.status_code, 503)
        data = response.json()
        self.assertEqual(data["status"], "unhealthy")
        db_check = next(check for check in data["checks"] if check["name"] == "database")
        self.assertFalse(db_check["status"])

    @pytest.mark.asyncio
    async def test_health_check_redis_unhealthy(self):
        """Verifica se o health check retorna erro quando o Redis está indisponível."""
        # Arrange
        self.redis_client_mock.is_healthy.return_value = False
        self.observability.setup_health_check_endpoint()
        
        # Act
        response = self.client.get("/health")
        
        # Assert
        self.assertEqual(response.status_code, 503)
        data = response.json()
        self.assertEqual(data["status"], "unhealthy")
        redis_check = next(check for check in data["checks"] if check["name"] == "redis_cache")
        self.assertFalse(redis_check["status"])

    def test_diagnostic_endpoint(self):
        """Verifica se o endpoint de diagnóstico retorna informações corretas."""
        # Arrange
        self.observability.setup_diagnostic_endpoint()
        
        # Act
        response = self.client.get("/diagnostic")
        
        # Assert
        self.assertEqual(response.status_code, 200)
        data = response.json()
        self.assertEqual(data["service"], "iam-audit-test")
        self.assertEqual(data["version"], "1.0.0")
        self.assertEqual(data["build"], "test-build")
        self.assertEqual(data["commit"], "abcdef123456")
        self.assertEqual(data["environment"], "test")
        self.assertEqual(data["region"], "us-east")

    @pytest.mark.asyncio
    async def test_startup_handler(self):
        """Verifica se o handler de startup atualiza as métricas corretamente."""
        # Arrange
        with patch("api.app.integrations.observability.AUDIT_SERVICE_INFO.labels") as mock_labels:
            mock_gauge = MagicMock()
            mock_labels.return_value = mock_gauge
            self.observability.setup_lifecycle_handlers()
            
            # Act
            await self.app.router.startup()
            
            # Assert
            mock_labels.assert_called_with(
                version="1.0.0", 
                build_id="test-build",
                commit_hash="abcdef123456", 
                environment="test",
                region="us-east"
            )
            mock_gauge.set.assert_called_with(1)

    @pytest.mark.asyncio
    async def test_shutdown_handler(self):
        """Verifica se o handler de shutdown atualiza as métricas corretamente."""
        # Arrange
        with patch("api.app.integrations.observability.AUDIT_SERVICE_INFO.labels") as mock_labels:
            mock_gauge = MagicMock()
            mock_labels.return_value = mock_gauge
            self.observability.setup_lifecycle_handlers()
            
            # Act
            await self.app.router.shutdown()
            
            # Assert
            mock_labels.assert_called_with(
                version="1.0.0", 
                build_id="test-build",
                commit_hash="abcdef123456", 
                environment="test",
                region="us-east"
            )
            mock_gauge.set.assert_called_with(0)

    def test_setup_health_verification_methods(self):
        """Verifica se os métodos de verificação de saúde são configurados corretamente."""
        # Arrange & Act
        self.observability.setup_health_verification_methods()
        
        # Assert
        self.assertTrue(callable(self.observability.verify_database_health))
        self.assertTrue(callable(self.observability.verify_cache_health))
        self.assertTrue(callable(self.observability.verify_queue_health))
        self.assertTrue(callable(self.observability.verify_storage_health))

    @pytest.mark.asyncio
    async def test_verify_database_health(self):
        """Verifica se a verificação de saúde do banco de dados funciona corretamente."""
        # Arrange
        self.observability.setup_health_verification_methods()
        
        # Act - Caso de sucesso
        result_success = await self.observability.verify_database_health()
        
        # Caso de falha
        self.db_client_mock.is_healthy.return_value = False
        result_failure = await self.observability.verify_database_health()
        
        # Caso de exceção
        self.db_client_mock.is_healthy.side_effect = Exception("DB connection failed")
        result_exception = await self.observability.verify_database_health()
        
        # Assert
        self.assertTrue(result_success["status"])
        self.assertEqual(result_success["name"], "database")
        
        self.assertFalse(result_failure["status"])
        self.assertEqual(result_failure["name"], "database")
        self.assertEqual(result_failure["message"], "Database connection check failed")
        
        self.assertFalse(result_exception["status"])
        self.assertEqual(result_exception["name"], "database")
        self.assertEqual(result_exception["message"], "Exception: DB connection failed")

    def test_convenience_function(self):
        """Verifica se a função de conveniência setup_observability funciona corretamente."""
        # Arrange
        app = FastAPI()
        
        # Act
        with patch("api.app.integrations.observability.ObservabilityIntegration") as mock_class:
            mock_instance = MagicMock()
            mock_class.return_value = mock_instance
            
            setup_observability(
                app=app,
                service_name="test-service",
                service_version="2.0.0",
                db_client=self.db_client_mock
            )
            
            # Assert
            mock_class.assert_called_once()
            mock_instance.setup_metrics.assert_called_once()
            mock_instance.setup_metrics_endpoint.assert_called_once()
            mock_instance.setup_health_check_endpoint.assert_called_once()
            mock_instance.setup_diagnostic_endpoint.assert_called_once()
            mock_instance.setup_health_verification_methods.assert_called_once()
            mock_instance.setup_lifecycle_handlers.assert_called_once()


if __name__ == "__main__":
    unittest.main()