"""
Testes unitários para o módulo de métricas do framework de observabilidade.

Estes testes verificam o correto funcionamento da criação, registro e uso
de métricas Prometheus, incluindo decoradores e manipulação de contexto multi-tenant.
"""

import unittest
from unittest import mock
import asyncio
from datetime import datetime

import pytest
from prometheus_client import CollectorRegistry, Counter, Histogram, Gauge

from src.observability.config import MetricsConfig
from src.observability.metrics import (
    MetricsManager,
    get_metrics_manager,
    set_metrics_manager,
    instrument_function,
    instrument_audit_event
)


class TestMetricsManager(unittest.TestCase):
    """Testes para o gerenciador de métricas."""

    def setUp(self):
        """Configuração para cada teste."""
        # Cria um registry isolado para cada teste
        self.registry = CollectorRegistry()
        self.config = MetricsConfig(
            namespace="test",
            subsystem="audit",
            enabled=True
        )
        self.default_labels = {
            "service": "iam-audit-service",
            "version": "1.0.0",
            "environment": "test"
        }
        self.metrics = MetricsManager(
            registry=self.registry,
            config=self.config,
            default_labels=self.default_labels
        )

    def test_create_counter(self):
        """Testa a criação de métricas tipo Counter."""
        counter = self.metrics.create_counter(
            name="test_counter",
            description="Counter para teste",
            labels=["tenant", "region"]
        )
        
        # Verifica se o counter foi criado corretamente
        self.assertIsInstance(counter, Counter)
        
        # Verifica se as labels estão corretas
        counter.labels(tenant="tenant1", region="us-east").inc()
        samples = list(counter.collect()[0].samples)
        
        # Verifica se o sample foi criado com as labels corretas
        sample = samples[0]
        self.assertEqual(sample.value, 1.0)
        self.assertEqual(sample.labels["tenant"], "tenant1")
        self.assertEqual(sample.labels["region"], "us-east")
        
        # Verifica se as default labels foram aplicadas
        for label, value in self.default_labels.items():
            self.assertEqual(sample.labels[label], value)

    def test_create_histogram(self):
        """Testa a criação de métricas tipo Histogram."""
        histogram = self.metrics.create_histogram(
            name="test_histogram",
            description="Histogram para teste",
            labels=["tenant", "region"],
            buckets=[0.01, 0.1, 1.0, 5.0]
        )
        
        # Verifica se o histogram foi criado corretamente
        self.assertIsInstance(histogram, Histogram)
        
        # Verifica se as labels estão corretas
        histogram.labels(tenant="tenant1", region="eu-west").observe(0.5)
        
        # Verifica se o nome está correto (namespace_subsystem_name)
        self.assertTrue(
            any("test_audit_test_histogram" in str(s.name) for s in histogram.collect()[0].samples)
        )

    def test_create_gauge(self):
        """Testa a criação de métricas tipo Gauge."""
        gauge = self.metrics.create_gauge(
            name="test_gauge",
            description="Gauge para teste",
            labels=["tenant", "region"]
        )
        
        # Verifica se o gauge foi criado corretamente
        self.assertIsInstance(gauge, Gauge)
        
        # Verifica se as labels estão corretas
        gauge.labels(tenant="tenant1", region="eu-west").set(42)
        samples = list(gauge.collect()[0].samples)
        
        # Verifica se o sample foi criado com as labels corretas
        sample = samples[0]
        self.assertEqual(sample.value, 42.0)
        self.assertEqual(sample.labels["tenant"], "tenant1")
        self.assertEqual(sample.labels["region"], "eu-west")

    def test_create_metrics_with_custom_namespace(self):
        """Testa a criação de métricas com namespace personalizado."""
        # Cria métrica com namespace customizado
        counter = self.metrics.create_counter(
            name="custom_counter",
            description="Counter com namespace personalizado",
            namespace="custom_ns",
            labels=["tenant"]
        )
        
        counter.labels(tenant="tenant1").inc(2)
        samples = list(counter.collect()[0].samples)
        
        # Verifica se o namespace está correto
        self.assertTrue(
            any("custom_ns_audit_custom_counter" in str(s.name) for s in counter.collect()[0].samples)
        )
        
        # Verifica se o valor está correto
        self.assertEqual(samples[0].value, 2.0)

    def test_singleton_management(self):
        """Testa o gerenciamento do singleton global de métricas."""
        # Limpa o singleton global
        set_metrics_manager(None)
        self.assertIsNone(get_metrics_manager())
        
        # Define um novo manager e verifica
        set_metrics_manager(self.metrics)
        self.assertEqual(get_metrics_manager(), self.metrics)
        
        # Tenta definir um valor inválido
        with self.assertRaises(ValueError):
            set_metrics_manager("not a metrics manager")


@pytest.mark.asyncio
class TestMetricsDecorators:
    """Testes para os decoradores de métricas."""

    @pytest.fixture
    async def setup_metrics(self):
        """Configura o gerenciador de métricas para testes."""
        registry = CollectorRegistry()
        config = MetricsConfig(
            namespace="test",
            subsystem="audit",
            enabled=True
        )
        default_labels = {
            "service": "iam-audit-service",
            "version": "1.0.0",
            "environment": "test"
        }
        metrics = MetricsManager(
            registry=registry,
            config=config,
            default_labels=default_labels
        )
        set_metrics_manager(metrics)
        
        # Cria métricas para teste
        metrics.http_request_duration_seconds = metrics.create_histogram(
            name="http_request_duration_seconds",
            description="Duração das requisições HTTP",
            labels=["method", "path", "tenant", "region"]
        )
        
        metrics.audit_events_total = metrics.create_counter(
            name="audit_events_total",
            description="Total de eventos de auditoria processados",
            labels=["event_type", "tenant", "region", "status"]
        )
        
        metrics.audit_events_errors_total = metrics.create_counter(
            name="audit_events_errors_total",
            description="Total de erros no processamento de eventos de auditoria",
            labels=["event_type", "tenant", "region", "error_type"]
        )
        
        metrics.audit_events_duration_seconds = metrics.create_histogram(
            name="audit_events_duration_seconds",
            description="Duração do processamento de eventos de auditoria",
            labels=["event_type", "tenant", "region"]
        )
        
        yield metrics
        set_metrics_manager(None)

    @pytest.mark.asyncio
    async def test_instrument_function_decorator(self, setup_metrics):
        """Testa o decorador instrument_function."""
        metrics = setup_metrics
        
        # Cria um histograma para teste
        histogram = metrics.create_histogram(
            name="test_function_duration",
            description="Duração de funções de teste",
            labels=["function", "tenant", "region"]
        )
        
        # Define uma função decorada para teste
        @instrument_function(
            histogram=histogram,
            labels={"function": "test_function"},
            extract_labels_from_args={"tenant": 0, "region": 1}
        )
        async def test_function(tenant, region, value):
            await asyncio.sleep(0.01)
            if value < 0:
                raise ValueError("Valor negativo não permitido")
            return value * 2
        
        # Testa execução bem-sucedida
        result = await test_function("tenant1", "eu-west", 5)
        assert result == 10
        
        # Testa captura de exceção
        with pytest.raises(ValueError):
            await test_function("tenant2", "us-east", -5)
        
        # Verifica se as métricas foram registradas
        samples = list(histogram.collect()[0].samples)
        tenants = set(s.labels.get("tenant") for s in samples if "tenant" in s.labels)
        assert "tenant1" in tenants
        assert "tenant2" in tenants

    @pytest.mark.asyncio
    async def test_instrument_audit_event_decorator(self, setup_metrics):
        """Testa o decorador instrument_audit_event."""
        metrics = setup_metrics
        
        # Define uma classe de evento para teste
        class TestEvent:
            def __init__(self, event_type, tenant_id, details=None):
                self.event_type = event_type
                self.tenant_id = tenant_id
                self.details = details or {}
        
        # Define uma função decorada para teste
        @instrument_audit_event(event_type="login")
        async def process_login_event(event, tenant, region):
            await asyncio.sleep(0.01)
            if event.details.get("error"):
                raise ValueError("Evento de login com erro")
            return {"status": "success", "event_type": event.event_type}
        
        # Testa execução bem-sucedida
        event_ok = TestEvent("login", "tenant1")
        result = await process_login_event(event_ok, "tenant1", "eu-west")
        assert result["status"] == "success"
        
        # Testa captura de exceção
        event_error = TestEvent("login", "tenant2", {"error": True})
        with pytest.raises(ValueError):
            await process_login_event(event_error, "tenant2", "us-east")
        
        # Verifica se as métricas foram registradas
        counter_samples = list(metrics.audit_events_total.collect()[0].samples)
        error_samples = list(metrics.audit_events_errors_total.collect()[0].samples)
        
        # Deve ter pelo menos uma métrica de evento processado
        assert len(counter_samples) > 0
        # Deve ter pelo menos uma métrica de erro
        assert len(error_samples) > 0
        
        # Verifica se o evento de erro foi registrado com o tenant correto
        error_tenant = next(
            (s.labels.get("tenant") for s in error_samples if s.labels.get("tenant") == "tenant2"),
            None
        )
        assert error_tenant == "tenant2"


if __name__ == "__main__":
    unittest.main()