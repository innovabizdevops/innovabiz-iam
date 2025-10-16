"""
Testes unitários para o sistema de observabilidade multi-camada da INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ - Sistema de Observabilidade Multi-Camada
Data: 20/08/2025
"""

import os
import json
import time
import logging
import unittest
from unittest.mock import patch, MagicMock, ANY
import pytest
from typing import Dict, Any

# Importar módulos de observabilidade
from observability.core import (
    MultiLayerObservabilityConfig,
    MultiLayerObservabilityConfigurator, 
    ObservabilityLevel,
    UnifiedMonitor,
    create_observer,
    trace_request,
    trace_auth,
    trace_transaction,
    trace_risk
)


class TestMultiLayerObservabilityConfig(unittest.TestCase):
    """Testes para a configuração de observabilidade multi-camada."""
    
    def test_default_values(self):
        """Testa os valores padrão da configuração."""
        config = MultiLayerObservabilityConfig()
        
        # Verificar valores padrão
        self.assertEqual(config.tenant_id, "default")
        self.assertEqual(config.observability_level, ObservabilityLevel.STANDARD)
        self.assertTrue(config.logging_enabled)
        self.assertEqual(config.log_level, "INFO")
        self.assertFalse(config.pii_masking_enabled)
    
    def test_copy_config(self):
        """Testa a cópia da configuração."""
        config1 = MultiLayerObservabilityConfig()
        config1.tenant_id = "tenant1"
        config1.module_name = "module1"
        
        # Copiar configuração
        config2 = config1.copy()
        
        # Verificar que é uma cópia independente
        self.assertEqual(config2.tenant_id, "tenant1")
        self.assertEqual(config2.module_name, "module1")
        
        # Modificar a cópia não afeta o original
        config2.tenant_id = "tenant2"
        self.assertEqual(config1.tenant_id, "tenant1")
        self.assertEqual(config2.tenant_id, "tenant2")


@pytest.fixture
def mock_env_vars():
    """Fixture para simular variáveis de ambiente para testes."""
    env_vars = {
        'TENANT_ID': 'test_tenant',
        'OTEL_SERVICE_NAME': 'test_service',
        'OTEL_ENVIRONMENT': 'test',
        'MARKET_CONTEXT': 'angola',
        'REGION': 'luanda',
        'LOG_LEVEL': 'DEBUG',
        'STRUCTURED_LOGGING': 'true',
        'TRACING_ENABLED': 'true',
        'METRICS_ENABLED': 'true',
        'PII_MASKING_ENABLED': 'true',
        'OBSERVABILITY_LEVEL': 'advanced'
    }
    
    with patch.dict(os.environ, env_vars):
        yield env_vars


@pytest.fixture
def mock_logger():
    """Fixture para criar um logger mock para testes."""
    logger = MagicMock()
    return logger


@pytest.fixture
def mock_otel():
    """Fixture para criar mocks do OpenTelemetry."""
    tracer_provider = MagicMock()
    tracer = MagicMock()
    tracer_provider.get_tracer.return_value = tracer
    
    meter_provider = MagicMock()
    meter = MagicMock()
    meter_provider.get_meter.return_value = meter
    
    with patch('observability.core.multi_layer_config.TracerProvider', return_value=tracer_provider), \
         patch('observability.core.multi_layer_config.MeterProvider', return_value=meter_provider):
        yield {
            'tracer_provider': tracer_provider,
            'tracer': tracer,
            'meter_provider': meter_provider,
            'meter': meter
        }


class TestMultiLayerObservabilityConfigurator:
    """Testes para o configurador de observabilidade multi-camada."""
    
    def test_from_env(self, mock_env_vars, mock_otel):
        """Testa a criação do configurador a partir de variáveis de ambiente."""
        configurator = MultiLayerObservabilityConfigurator.from_env()
        
        # Verificar se as variáveis de ambiente foram carregadas corretamente
        assert configurator.config.tenant_id == 'test_tenant'
        assert configurator.config.module_name == 'test_service'
        assert configurator.config.environment == 'test'
        assert configurator.config.market_context == 'angola'
        assert configurator.config.region == 'luanda'
        assert configurator.config.log_level == 'DEBUG'
        assert configurator.config.structured_logging == True
        assert configurator.config.tracing_enabled == True
        assert configurator.config.metrics_enabled == True
        assert configurator.config.pii_masking_enabled == True
        assert configurator.config.observability_level == ObservabilityLevel.ADVANCED


class TestUnifiedMonitor:
    """Testes para o monitor unificado de observabilidade."""
    
    def test_create_monitor(self, mock_env_vars, mock_logger, mock_otel):
        """Testa a criação do monitor unificado."""
        configurator = MultiLayerObservabilityConfigurator(
            MultiLayerObservabilityConfig(), "test.logger"
        )
        configurator.logger = mock_logger
        configurator.tracer = mock_otel['tracer']
        configurator.meter = mock_otel['meter']
        
        monitor = UnifiedMonitor(configurator)
        
        # Verificar se o monitor foi criado corretamente
        assert monitor.base_monitor is not None
        assert monitor.security is not None
        assert monitor.business is not None


@patch('observability.core.create_monitor_from_env')
class TestDecorators:
    """Testes para os decoradores de conveniência."""
    
    def test_trace_request(self, mock_create_monitor):
        """Testa o decorador trace_request."""
        monitor_mock = MagicMock()
        base_monitor_mock = MagicMock()
        monitor_mock.base_monitor = base_monitor_mock
        mock_create_monitor.return_value = monitor_mock
        
        # Chamar o decorador
        decorator = trace_request("test_operation")
        
        # Verificar se o monitor foi criado e o decorador foi chamado corretamente
        mock_create_monitor.assert_called_once()
        base_monitor_mock.trace_request.assert_called_once_with("test_operation")
    
    def test_trace_auth(self, mock_create_monitor):
        """Testa o decorador trace_auth."""
        monitor_mock = MagicMock()
        security_mock = MagicMock()
        monitor_mock.security = security_mock
        mock_create_monitor.return_value = monitor_mock
        
        # Chamar o decorador
        decorator = trace_auth("mfa")
        
        # Verificar se o monitor foi criado e o decorador foi chamado corretamente
        mock_create_monitor.assert_called_once()
        security_mock.trace_authentication.assert_called_once_with(auth_type="mfa")
    
    def test_trace_transaction(self, mock_create_monitor):
        """Testa o decorador trace_transaction."""
        monitor_mock = MagicMock()
        business_mock = MagicMock()
        monitor_mock.business = business_mock
        mock_create_monitor.return_value = monitor_mock
        
        # Chamar o decorador
        decorator = trace_transaction("payment")
        
        # Verificar se o monitor foi criado e o decorador foi chamado corretamente
        mock_create_monitor.assert_called_once()
        business_mock.trace_business_transaction.assert_called_once_with(transaction_type="payment")
    
    def test_trace_risk(self, mock_create_monitor):
        """Testa o decorador trace_risk."""
        monitor_mock = MagicMock()
        security_mock = MagicMock()
        monitor_mock.security = security_mock
        mock_create_monitor.return_value = monitor_mock
        
        # Chamar o decorador
        decorator = trace_risk("transaction")
        
        # Verificar se o monitor foi criado e o decorador foi chamado corretamente
        mock_create_monitor.assert_called_once()
        security_mock.trace_risk_assessment.assert_called_once_with(risk_type="transaction")


class TestIntegrationFlow:
    """Testes de integração para fluxos comuns de observabilidade."""
    
    @patch('observability.core.multi_layer_monitor.time')
    def test_security_authentication_flow(self, mock_time, mock_env_vars, mock_logger, mock_otel):
        """Testa o fluxo completo de autenticação com observabilidade."""
        # Configurar mocks
        mock_time.time.return_value = 1000.0
        mock_span = MagicMock()
        mock_otel['tracer'].start_span.return_value = mock_span
        mock_otel['tracer'].start_as_current_span = MagicMock(return_value=mock_span)
        
        # Criar monitor
        configurator = MultiLayerObservabilityConfigurator.from_env()
        configurator.logger = mock_logger
        configurator.tracer = mock_otel['tracer']
        configurator.meter = mock_otel['meter']
        
        monitor = UnifiedMonitor(configurator)
        
        # Criar função a ser decorada
        @monitor.security.trace_authentication(auth_type="password")
        def authenticate(username, password):
            return {"success": True, "user_id": "123"}
        
        # Chamar função
        result = authenticate("testuser", "password123")
        
        # Verificar resultado
        assert result["success"] == True
        
        # Verificar chamadas de telemetria
        mock_logger.info.assert_any_call(
            ANY, 
            extra={"event_category": "SECURITY", "event_severity": "INFO"}
        )
        mock_otel['tracer'].start_span.assert_called_with(
            "auth.password", 
            attributes=ANY
        )
        # Verificar contadores de autenticação
        mock_counter_calls = [
            call for name, args, kwargs in mock_otel['meter'].create_counter.mock_calls 
            if 'auth' in str(args)
        ]
        assert len(mock_counter_calls) >= 1
    
    @patch('observability.core.multi_layer_monitor.time')
    def test_business_transaction_flow(self, mock_time, mock_env_vars, mock_logger, mock_otel):
        """Testa o fluxo completo de transação com observabilidade."""
        # Configurar mocks
        mock_time.time.return_value = 1000.0
        mock_span = MagicMock()
        mock_otel['tracer'].start_span.return_value = mock_span
        mock_otel['tracer'].start_as_current_span = MagicMock(return_value=mock_span)
        
        # Criar monitor
        configurator = MultiLayerObservabilityConfigurator.from_env()
        configurator.logger = mock_logger
        configurator.tracer = mock_otel['tracer']
        configurator.meter = mock_otel['meter']
        
        monitor = UnifiedMonitor(configurator)
        
        # Criar função a ser decorada
        @monitor.business.trace_business_transaction(transaction_type="payment")
        def process_payment(transaction_id, amount, currency="USD"):
            return {"status": "SUCCESS", "transaction_id": transaction_id}
        
        # Chamar função
        result = process_payment("tx123", 100.0, currency="USD")
        
        # Verificar resultado
        assert result["status"] == "SUCCESS"
        
        # Verificar chamadas de telemetria
        mock_logger.info.assert_any_call(
            ANY, 
            extra={"event_category": "BUSINESS", "event_severity": "INFO"}
        )
        mock_otel['tracer'].start_span.assert_called_with(
            "business.transaction.payment", 
            attributes=ANY
        )
        # Verificar métricas de transação
        mock_histogram_calls = [
            call for name, args, kwargs in mock_otel['meter'].create_histogram.mock_calls 
            if 'transaction' in str(args)
        ]
        assert len(mock_histogram_calls) >= 1