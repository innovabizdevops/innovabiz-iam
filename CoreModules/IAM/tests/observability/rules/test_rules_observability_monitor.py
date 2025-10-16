"""
Testes unitários para o monitor de observabilidade do sistema de regras.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/Observabilidade
Data: 21/08/2025
"""

import time
import unittest
from unittest import mock
from functools import wraps

import pytest
from opentelemetry import trace, metrics
from opentelemetry.trace import Span, SpanKind
from opentelemetry.metrics import Counter, Histogram

# Importar módulos a serem testados
from observability.rules.rules_observability_config import ObservabilityConfig, RulesObservabilityConfigurator
from observability.rules.rules_observability_monitor import RulesObservabilityMonitor, instrument_connector


class TestRulesObservabilityMonitor(unittest.TestCase):
    """Testes unitários para a classe RulesObservabilityMonitor."""
    
    def setUp(self):
        """Configura ambiente para testes."""
        # Mock para o configurador
        self.mock_configurator = mock.MagicMock(spec=RulesObservabilityConfigurator)
        
        # Mock para logger, tracer e meter
        self.mock_logger = mock.MagicMock()
        self.mock_tracer = mock.MagicMock()
        self.mock_meter = mock.MagicMock()
        
        # Mock para span
        self.mock_span = mock.MagicMock(spec=Span)
        self.mock_span.__enter__ = mock.MagicMock(return_value=self.mock_span)
        self.mock_span.__exit__ = mock.MagicMock(return_value=None)
        self.mock_tracer.start_as_current_span.return_value = self.mock_span
        
        # Mock para métricas
        self.mock_counter = mock.MagicMock(spec=Counter)
        self.mock_histogram = mock.MagicMock(spec=Histogram)
        self.mock_meter.create_counter.return_value = self.mock_counter
        self.mock_meter.create_histogram.return_value = self.mock_histogram
        
        # Configurar mocks no configurador
        self.mock_configurator.setup_logger.return_value = self.mock_logger
        self.mock_configurator.setup_tracer.return_value = self.mock_tracer
        self.mock_configurator.setup_meter.return_value = self.mock_meter
        
        # Criar monitor
        self.component_name = "test-component"
        self.monitor = RulesObservabilityMonitor(self.mock_configurator, self.component_name)
    
    def test_init(self):
        """Testa inicialização do monitor."""
        # Verificar configuração
        self.assertEqual(self.monitor.configurator, self.mock_configurator)
        self.assertEqual(self.monitor.component_name, self.component_name)
        self.assertEqual(self.monitor.logger, self.mock_logger)
        self.assertEqual(self.monitor.tracer, self.mock_tracer)
        self.assertEqual(self.monitor.meter, self.mock_meter)
        
        # Verificar chamadas para setup
        self.mock_configurator.setup_logger.assert_called_once()
        self.mock_configurator.setup_tracer.assert_called_once()
        self.mock_configurator.setup_meter.assert_called_once()
        
        # Verificar criação de métricas
        self.assertEqual(self.mock_meter.create_counter.call_count, 8)
        self.assertEqual(self.mock_meter.create_histogram.call_count, 3)
    
    @mock.patch("time.time")
    async def test_trace_rule_evaluation(self, mock_time):
        """Testa decorador para avaliação de regras."""
        # Configurar mocks
        mock_time.side_effect = [100.0, 100.5]  # Início e fim da execução
        
        # Criar função mock para decorar
        @self.monitor.trace_rule_evaluation
        async def mock_evaluate_rules(self, ruleset, event):
            return {"rule1": mock.MagicMock(triggered=True), "rule2": mock.MagicMock(triggered=False)}
        
        # Criar mocks para argumentos
        mock_self = mock.MagicMock()
        mock_ruleset = mock.MagicMock()
        mock_ruleset.id = "test-ruleset"
        mock_ruleset.rules = ["rule1", "rule2"]
        mock_event = mock.MagicMock()
        mock_event.type = "test-event"
        
        # Executar função decorada
        result = await mock_evaluate_rules(mock_self, mock_ruleset, mock_event)
        
        # Verificar chamadas de tracing
        self.mock_tracer.start_as_current_span.assert_called_once_with(
            name=f"{self.component_name}.evaluate_rules",
            kind=SpanKind.INTERNAL,
            attributes={
                "ruleset.id": "test-ruleset",
                "ruleset.num_rules": 2,
                "event.type": "test-event",
            },
        )
        
        # Verificar atributos do span
        self.mock_span.set_attribute.assert_any_call("rules.total", 2)
        self.mock_span.set_attribute.assert_any_call("rules.triggered", 1)
        self.mock_span.set_attribute.assert_any_call("rules.evaluation.duration_ms", 500.0)
        
        # Verificar métricas
        self.mock_counter.add.assert_any_call(1, {"ruleset_id": "test-ruleset"})
        self.mock_counter.add.assert_any_call(1, {"ruleset_id": "test-ruleset"})
        self.mock_histogram.record.assert_called_once_with(500.0, {"ruleset_id": "test-ruleset"})
        
        # Verificar log
        self.mock_logger.info.assert_called_once()
        
        # Verificar resultado
        self.assertEqual(len(result), 2)
    
    @mock.patch("time.time")
    async def test_trace_bureau_request(self, mock_time):
        """Testa decorador para requisições ao Bureau de Créditos."""
        # Configurar mocks
        mock_time.side_effect = [100.0, 100.3]  # Início e fim da execução
        
        # Criar função mock para decorar
        @self.monitor.trace_bureau_request
        async def mock_bureau_request(self, **kwargs):
            return {"score": 750}
        
        # Criar mock para self
        mock_self = mock.MagicMock()
        
        # Executar função decorada
        result = await mock_bureau_request(
            mock_self,
            provider="test-provider",
            data_type="credit-score",
            document_id="123456789",
        )
        
        # Verificar chamadas de tracing
        self.mock_tracer.start_as_current_span.assert_called_once_with(
            name=f"{self.component_name}.bureau_request",
            kind=SpanKind.CLIENT,
            attributes={
                "bureau.provider": "test-provider",
                "bureau.data_type": "credit-score",
                "bureau.document_id": "123456789",
            },
        )
        
        # Verificar atributos do span
        self.mock_span.set_attribute.assert_any_call("bureau.response.duration_ms", 300.0)
        self.mock_span.set_attribute.assert_any_call("bureau.response.status", "success")
        
        # Verificar métricas
        self.mock_counter.add.assert_called_once_with(1, {
            "provider": "test-provider",
            "data_type": "credit-score",
        })
        self.mock_histogram.record.assert_called_once_with(300.0, {
            "provider": "test-provider",
            "data_type": "credit-score",
        })
        
        # Verificar log
        self.mock_logger.info.assert_called_once()
        
        # Verificar resultado
        self.assertEqual(result, {"score": 750})
    
    @mock.patch("time.time")
    async def test_trace_trustguard_request(self, mock_time):
        """Testa decorador para requisições ao TrustGuard."""
        # Configurar mocks
        mock_time.side_effect = [100.0, 100.2]  # Início e fim da execução
        
        # Criar função mock para decorar
        @self.monitor.trace_trustguard_request
        async def _make_request(self, method, endpoint, data=None):
            return {"status": "success"}
        
        # Criar mock para self
        mock_self = mock.MagicMock()
        
        # Executar função decorada
        result = await _make_request(
            mock_self,
            method="POST",
            endpoint="/api/access/evaluate",
            data={"user_id": "user123"},
        )
        
        # Verificar chamadas de tracing
        self.mock_tracer.start_as_current_span.assert_called_once_with(
            name=f"{self.component_name}.trustguard_request",
            kind=SpanKind.CLIENT,
            attributes={
                "trustguard.endpoint": "/api/access/evaluate",
                "trustguard.user_id": "unknown",
            },
        )
        
        # Verificar atributos do span
        self.mock_span.set_attribute.assert_any_call("trustguard.response.duration_ms", 200.0)
        self.mock_span.set_attribute.assert_any_call("trustguard.response.status", "success")
        
        # Verificar métricas
        self.mock_counter.add.assert_called_once_with(1, {
            "endpoint": "/api/access/evaluate",
            "method": "_make_request",
        })
        self.mock_histogram.record.assert_called_once_with(200.0, {
            "endpoint": "/api/access/evaluate",
            "method": "_make_request",
        })
        
        # Verificar log
        self.mock_logger.info.assert_called_once()
        
        # Verificar resultado
        self.assertEqual(result, {"status": "success"})
    
    @mock.patch("time.time")
    async def test_trace_access_evaluation(self, mock_time):
        """Testa decorador para avaliações de acesso."""
        # Configurar mocks
        mock_time.side_effect = [100.0, 100.4]  # Início e fim da execução
        
        # Criar função mock para decorar
        @self.monitor.trace_access_evaluation
        async def evaluate_access(self, request):
            return mock.MagicMock(decision="ALLOW", risk_level="LOW")
        
        # Criar mocks para argumentos
        mock_self = mock.MagicMock()
        mock_request = mock.MagicMock()
        mock_request.request_id = "req123"
        mock_request.user_context = mock.MagicMock(user_id="user123")
        mock_request.resource_context = mock.MagicMock(
            resource_id="res456",
            resource_type="document",
            action="read",
        )
        
        # Executar função decorada
        result = await evaluate_access(mock_self, mock_request)
        
        # Verificar chamadas de tracing
        self.mock_tracer.start_as_current_span.assert_called_once_with(
            name=f"{self.component_name}.evaluate_access",
            kind=SpanKind.INTERNAL,
            attributes={
                "access.request_id": "req123",
                "access.user_id": "user123",
                "access.resource_id": "res456",
                "access.resource_type": "document",
                "access.action": "read",
            },
        )
        
        # Verificar atributos do span
        self.mock_span.set_attribute.assert_any_call("access.decision", "ALLOW")
        self.mock_span.set_attribute.assert_any_call("access.risk_level", "LOW")
        self.mock_span.set_attribute.assert_any_call("access.response.duration_ms", 400.0)
        
        # Verificar métricas
        self.mock_counter.add.assert_called_once_with(1, {
            "decision": "ALLOW",
            "risk_level": "LOW",
            "resource_type": "document",
            "action": "read",
        })
        self.mock_histogram.record.assert_called_once_with(400.0, {
            "method": "evaluate_access",
        })
        
        # Verificar log
        self.mock_logger.info.assert_called_once()
        
        # Verificar resultado
        self.assertEqual(result.decision, "ALLOW")
        self.assertEqual(result.risk_level, "LOW")
    
    @mock.patch("time.time")
    def test_trace_cache_operation(self, mock_time):
        """Testa decorador para operações de cache."""
        # Configurar mocks
        mock_time.side_effect = [100.0, 100.1]  # Início e fim da execução
        
        # Criar função mock para decorar
        @self.monitor.trace_cache_operation
        def _get_from_cache(self, key):
            return (True, {"decision": "ALLOW"})
        
        # Criar mock para self
        mock_self = mock.MagicMock()
        
        # Executar função decorada
        hit, data = _get_from_cache(mock_self, "cache-key")
        
        # Verificar chamadas de tracing
        self.mock_tracer.start_as_current_span.assert_called_once_with(
            name=f"{self.component_name}.cache_operation",
            kind=SpanKind.INTERNAL,
            attributes={
                "cache.key": "cache-key",
                "cache.operation": "_get_from_cache",
            },
        )
        
        # Verificar atributos do span
        self.mock_span.set_attribute.assert_any_call("cache.operation.duration_ms", 100.0)
        self.mock_span.set_attribute.assert_any_call("cache.hit", True)
        
        # Verificar métricas para hit
        self.mock_counter.add.assert_called_once_with(1)
        
        # Verificar resultado
        self.assertTrue(hit)
        self.assertEqual(data, {"decision": "ALLOW"})
    
    def test_instrument_connector(self):
        """Testa instrumentação de conector."""
        # Criar mock para conector
        mock_connector = mock.MagicMock()
        original_make_request = mock_connector._make_request
        original_evaluate_access = mock_connector.evaluate_access
        original_get_from_cache = mock_connector._get_from_cache
        original_set_cache = mock_connector._set_cache
        
        # Instrumentar conector
        instrumented_connector = instrument_connector(mock_connector, self.monitor)
        
        # Verificar que métodos foram modificados
        self.assertNotEqual(instrumented_connector._make_request, original_make_request)
        self.assertNotEqual(instrumented_connector.evaluate_access, original_evaluate_access)
        self.assertNotEqual(instrumented_connector._get_from_cache, original_get_from_cache)
        self.assertNotEqual(instrumented_connector._set_cache, original_set_cache)
        
        # Verificar que é o mesmo objeto
        self.assertEqual(instrumented_connector, mock_connector)


if __name__ == "__main__":
    unittest.main()