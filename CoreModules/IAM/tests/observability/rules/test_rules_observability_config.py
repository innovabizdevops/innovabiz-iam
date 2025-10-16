"""
Testes unitários para a configuração de observabilidade do sistema de regras.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/Observabilidade
Data: 21/08/2025
"""

import os
import unittest
from unittest import mock

import pytest
from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.metrics import MeterProvider

# Importar módulo a ser testado
from observability.rules.rules_observability_config import ObservabilityConfig, RulesObservabilityConfigurator


class TestObservabilityConfig(unittest.TestCase):
    """Testes unitários para a classe ObservabilityConfig."""
    
    def test_default_config(self):
        """Testa criação de configuração com valores padrão."""
        config = ObservabilityConfig()
        
        # Verificar valores padrão
        self.assertEqual(config.service_name, "innovabiz-iam-rules")
        self.assertEqual(config.environment, "development")
        self.assertEqual(config.version, "1.0.0")
        self.assertEqual(config.tenant_id, "default")
        self.assertEqual(config.log_level, "INFO")
        self.assertTrue(config.tracing_enabled)
        self.assertTrue(config.metrics_enabled)
        self.assertEqual(config.metrics_export_interval_ms, 60000)
        self.assertDictEqual(config.tags, {})
    
    def test_custom_config(self):
        """Testa criação de configuração com valores customizados."""
        config = ObservabilityConfig(
            service_name="custom-service",
            environment="production",
            version="2.0.0",
            tenant_id="tenant1",
            log_level="DEBUG",
            tracing_enabled=False,
            metrics_enabled=False,
            metrics_export_interval_ms=30000,
            tags={"key1": "value1", "key2": "value2"},
        )
        
        # Verificar valores customizados
        self.assertEqual(config.service_name, "custom-service")
        self.assertEqual(config.environment, "production")
        self.assertEqual(config.version, "2.0.0")
        self.assertEqual(config.tenant_id, "tenant1")
        self.assertEqual(config.log_level, "DEBUG")
        self.assertFalse(config.tracing_enabled)
        self.assertFalse(config.metrics_enabled)
        self.assertEqual(config.metrics_export_interval_ms, 30000)
        self.assertDictEqual(config.tags, {"key1": "value1", "key2": "value2"})


class TestRulesObservabilityConfigurator(unittest.TestCase):
    """Testes unitários para a classe RulesObservabilityConfigurator."""
    
    def setUp(self):
        """Configura ambiente para testes."""
        self.config = ObservabilityConfig(
            service_name="test-service",
            environment="test",
            version="1.0.0",
            tenant_id="test-tenant",
            log_level="INFO",
            tracing_enabled=True,
            metrics_enabled=True,
        )
        self.logger_name = "test.logger"
        
        # Mock para trace e metrics
        self.trace_patcher = mock.patch("observability.rules.rules_observability_config.trace")
        self.metrics_patcher = mock.patch("observability.rules.rules_observability_config.metrics")
        
        self.mock_trace = self.trace_patcher.start()
        self.mock_metrics = self.metrics_patcher.start()
        
        # Mock para logging
        self.logging_patcher = mock.patch("observability.rules.rules_observability_config.logging")
        self.mock_logging = self.logging_patcher.start()
        
    def tearDown(self):
        """Limpa recursos após testes."""
        self.trace_patcher.stop()
        self.metrics_patcher.stop()
        self.logging_patcher.stop()
    
    def test_init_configurator(self):
        """Testa inicialização do configurador."""
        configurator = RulesObservabilityConfigurator(self.config, self.logger_name)
        
        # Verificar atributos
        self.assertEqual(configurator.config, self.config)
        self.assertEqual(configurator.logger_name, self.logger_name)
        self.assertIsNone(configurator._logger)
        self.assertIsNone(configurator._tracer)
        self.assertIsNone(configurator._meter)
        
        # Verificar recursos
        self.assertIsNotNone(configurator.resource)
    
    def test_setup_logger(self):
        """Testa configuração de logger."""
        # Mock para logger
        mock_logger = mock.MagicMock()
        self.mock_logging.getLogger.return_value = mock_logger
        
        # Mock para handler
        mock_console_handler = mock.MagicMock()
        mock_file_handler = mock.MagicMock()
        self.mock_logging.StreamHandler.return_value = mock_console_handler
        self.mock_logging.FileHandler.return_value = mock_file_handler
        
        # Mock para formatter
        mock_formatter = mock.MagicMock()
        self.mock_logging.Formatter.return_value = mock_formatter
        
        # Configurar logger
        configurator = RulesObservabilityConfigurator(self.config, self.logger_name)
        logger = configurator.setup_logger()
        
        # Verificar chamadas
        self.mock_logging.getLogger.assert_called_with(self.logger_name)
        self.mock_logging.StreamHandler.assert_called_once()
        mock_console_handler.setFormatter.assert_called_with(mock_formatter)
        mock_logger.addHandler.assert_called_with(mock_console_handler)
        mock_logger.addFilter.assert_called_once()
        
        # Verificar resultado
        self.assertEqual(logger, mock_logger)
        self.assertEqual(configurator._logger, mock_logger)
    
    def test_setup_tracer(self):
        """Testa configuração de tracer."""
        # Mock para tracer
        mock_tracer = mock.MagicMock()
        self.mock_trace.get_tracer.return_value = mock_tracer
        
        # Mock para provider e exporter
        mock_provider = mock.MagicMock()
        mock_exporter = mock.MagicMock()
        mock_processor = mock.MagicMock()
        
        with mock.patch("observability.rules.rules_observability_config.TracerProvider", return_value=mock_provider):
            with mock.patch("observability.rules.rules_observability_config.OTLPSpanExporter", return_value=mock_exporter):
                with mock.patch("observability.rules.rules_observability_config.BatchSpanProcessor", return_value=mock_processor):
                    # Configurar tracer
                    configurator = RulesObservabilityConfigurator(self.config, self.logger_name)
                    tracer = configurator.setup_tracer()
                    
                    # Verificar chamadas
                    mock_provider.add_span_processor.assert_called_with(mock_processor)
                    self.mock_trace.set_tracer_provider.assert_called_with(mock_provider)
                    self.mock_trace.get_tracer.assert_called_with(self.config.service_name, self.config.version)
                    
                    # Verificar resultado
                    self.assertEqual(tracer, mock_tracer)
                    self.assertEqual(configurator._tracer, mock_tracer)
    
    def test_setup_meter(self):
        """Testa configuração de meter."""
        # Mock para meter
        mock_meter = mock.MagicMock()
        self.mock_metrics.get_meter.return_value = mock_meter
        
        # Mock para provider e exporter
        mock_provider = mock.MagicMock()
        mock_exporter = mock.MagicMock()
        mock_reader = mock.MagicMock()
        
        with mock.patch("observability.rules.rules_observability_config.MeterProvider", return_value=mock_provider):
            with mock.patch("observability.rules.rules_observability_config.OTLPMetricsExporter", return_value=mock_exporter):
                with mock.patch("observability.rules.rules_observability_config.PeriodicExportingMetricReader", return_value=mock_reader):
                    # Configurar meter
                    configurator = RulesObservabilityConfigurator(self.config, self.logger_name)
                    meter = configurator.setup_meter()
                    
                    # Verificar chamadas
                    self.mock_metrics.set_meter_provider.assert_called_with(mock_provider)
                    self.mock_metrics.get_meter.assert_called_with(self.config.service_name, self.config.version)
                    
                    # Verificar resultado
                    self.assertEqual(meter, mock_meter)
                    self.assertEqual(configurator._meter, mock_meter)
    
    @mock.patch.dict(os.environ, {
        "OTEL_SERVICE_NAME": "env-service",
        "OTEL_ENVIRONMENT": "env-test",
        "OTEL_VERSION": "2.0.0",
        "TENANT_ID": "env-tenant",
        "LOG_LEVEL": "DEBUG",
        "TRACING_ENABLED": "false",
        "METRICS_ENABLED": "false",
        "OTEL_TAG_environment": "test",
        "OTEL_TAG_team": "iam",
    })
    def test_from_env(self):
        """Testa criação de configurador a partir de variáveis de ambiente."""
        configurator = RulesObservabilityConfigurator.from_env("env.logger")
        
        # Verificar configuração
        self.assertEqual(configurator.config.service_name, "env-service")
        self.assertEqual(configurator.config.environment, "env-test")
        self.assertEqual(configurator.config.version, "2.0.0")
        self.assertEqual(configurator.config.tenant_id, "env-tenant")
        self.assertEqual(configurator.config.log_level, "DEBUG")
        self.assertFalse(configurator.config.tracing_enabled)
        self.assertFalse(configurator.config.metrics_enabled)
        self.assertEqual(configurator.logger_name, "env.logger")
        
        # Verificar tags
        self.assertEqual(configurator.config.tags["environment"], "test")
        self.assertEqual(configurator.config.tags["team"], "iam")


if __name__ == "__main__":
    unittest.main()