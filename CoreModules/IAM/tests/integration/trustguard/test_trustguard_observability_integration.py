"""
Teste de integração para o sistema de observabilidade com TrustGuard.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
import unittest
import asyncio
from unittest import mock

import pytest
import httpx
from pydantic import BaseModel

# Importar módulos a serem testados
from integration.bureau-credito.trustguard.trust_guard_connector import TrustGuardConnector, AccessRequest, AccessDecision
from integration.bureau-credito.trustguard.trust_guard_observability_factory import TrustGuardObservabilityFactory
from observability.rules.rules_observability_config import ObservabilityConfig, RulesObservabilityConfigurator
from observability.rules.rules_observability_monitor import RulesObservabilityMonitor


class TestTrustGuardObservabilityIntegration:
    """Testes de integração para o sistema de observabilidade com TrustGuard."""
    
    @pytest.fixture(autouse=True)
    def setup_environment(self):
        """Configura variáveis de ambiente para testes."""
        # Variáveis para TrustGuard
        os.environ["TRUST_GUARD_API_URL"] = "http://mock-trustguard:8080/api"
        os.environ["TRUST_GUARD_API_KEY"] = "test-api-key"
        os.environ["TRUST_GUARD_TENANT_ID"] = "test-tenant"
        os.environ["TRUST_GUARD_TIMEOUT"] = "2.0"
        os.environ["TRUST_GUARD_CACHE_TTL"] = "60"
        os.environ["TRUST_GUARD_CACHE_ENABLED"] = "true"
        
        # Variáveis para observabilidade
        os.environ["OTEL_SERVICE_NAME"] = "test-trustguard-service"
        os.environ["OTEL_ENVIRONMENT"] = "test"
        os.environ["OTEL_VERSION"] = "1.0.0"
        os.environ["TENANT_ID"] = "test-tenant"
        os.environ["LOG_LEVEL"] = "INFO"
        os.environ["TRACING_ENABLED"] = "true"
        os.environ["METRICS_ENABLED"] = "true"
        os.environ["OTEL_TAG_component"] = "trustguard"
        os.environ["OTEL_TAG_test"] = "true"
        
        yield
        
        # Limpar variáveis após testes
        for var in [
            "TRUST_GUARD_API_URL", "TRUST_GUARD_API_KEY", "TRUST_GUARD_TENANT_ID",
            "TRUST_GUARD_TIMEOUT", "TRUST_GUARD_CACHE_TTL", "TRUST_GUARD_CACHE_ENABLED",
            "OTEL_SERVICE_NAME", "OTEL_ENVIRONMENT", "OTEL_VERSION", "TENANT_ID",
            "LOG_LEVEL", "TRACING_ENABLED", "METRICS_ENABLED", "OTEL_TAG_component",
            "OTEL_TAG_test"
        ]:
            if var in os.environ:
                del os.environ[var]
    
    @pytest.mark.asyncio
    async def test_create_monitored_connector(self):
        """Testa criação de conector monitorado."""
        # Mock para cliente HTTP e respostas
        mock_client = mock.AsyncMock(spec=httpx.AsyncClient)
        mock_response = mock.AsyncMock(spec=httpx.Response)
        mock_response.status_code = 200
        mock_response.json.return_value = {"status": "success", "decision": "ALLOW", "risk_level": "LOW"}
        mock_client.request.return_value = mock_response
        
        # Patch para cliente HTTP
        with mock.patch("integration.bureau-credito.trustguard.trust_guard_connector.httpx.AsyncClient", 
                       return_value=mock_client):
            # Criar conector monitorado
            connector = await TrustGuardObservabilityFactory.create_monitored_connector_async(
                component_name="test-component",
                logger_name="test.logger",
            )
            
            # Verificar tipo do conector
            assert isinstance(connector, TrustGuardConnector)
            
            # Criar request para teste
            request = AccessRequest(
                request_id="test-req-001",
                user_context={"user_id": "user123", "groups": ["users", "customers"]},
                resource_context={"resource_id": "doc456", "resource_type": "document", "action": "read"},
                environment_context={"ip": "192.168.1.1", "device": "mobile", "location": "br"},
                tenant_id="test-tenant",
            )
            
            # Testar avaliação de acesso com observabilidade
            result = await connector.evaluate_access(request)
            
            # Verificar resultado
            assert result.decision == "ALLOW"
            assert result.risk_level == "LOW"
            
            # Verificar chamada HTTP com headers corretos
            mock_client.request.assert_called_once()
            call_args = mock_client.request.call_args[1]
            assert "headers" in call_args
            assert call_args["headers"]["X-API-Key"] == "test-api-key"
            assert call_args["headers"]["X-Tenant-ID"] == "test-tenant"
    
    @pytest.mark.asyncio
    async def test_observability_error_handling(self):
        """Testa manipulação de erros com observabilidade."""
        # Mock para cliente HTTP e respostas
        mock_client = mock.AsyncMock(spec=httpx.AsyncClient)
        mock_client.request.side_effect = httpx.RequestError("Connection error")
        
        # Mock para logger
        mock_logger = mock.MagicMock()
        
        # Patch para cliente HTTP e logger
        with mock.patch("integration.bureau-credito.trustguard.trust_guard_connector.httpx.AsyncClient", 
                       return_value=mock_client):
            with mock.patch("observability.rules.rules_observability_config.logging.getLogger",
                          return_value=mock_logger):
                # Criar conector monitorado
                connector = await TrustGuardObservabilityFactory.create_monitored_connector_async(
                    component_name="test-component",
                    logger_name="test.logger",
                )
                
                # Criar request para teste
                request = AccessRequest(
                    request_id="test-req-002",
                    user_context={"user_id": "user456", "groups": ["users"]},
                    resource_context={"resource_id": "doc789", "resource_type": "document", "action": "write"},
                    environment_context={"ip": "10.0.0.1", "device": "desktop"},
                    tenant_id="test-tenant",
                )
                
                # Testar avaliação de acesso com erro
                with pytest.raises(httpx.RequestError):
                    await connector.evaluate_access(request)
                
                # Verificar log de erro
                mock_logger.error.assert_called()
    
    @pytest.mark.asyncio
    async def test_cache_observability(self):
        """Testa observabilidade para operações de cache."""
        # Mock para cliente HTTP e respostas
        mock_client = mock.AsyncMock(spec=httpx.AsyncClient)
        mock_response = mock.AsyncMock(spec=httpx.Response)
        mock_response.status_code = 200
        mock_response.json.return_value = {"status": "success", "decision": "ALLOW", "risk_level": "LOW"}
        mock_client.request.return_value = mock_response
        
        # Mock para métricas
        mock_meter = mock.MagicMock()
        mock_counter = mock.MagicMock()
        mock_meter.create_counter.return_value = mock_counter
        
        # Patches
        with mock.patch("integration.bureau-credito.trustguard.trust_guard_connector.httpx.AsyncClient", 
                       return_value=mock_client):
            with mock.patch("observability.rules.rules_observability_monitor.metrics.get_meter",
                          return_value=mock_meter):
                # Criar conector monitorado
                connector = await TrustGuardObservabilityFactory.create_monitored_connector_async(
                    component_name="test-component",
                    logger_name="test.logger",
                )
                
                # Ativar cache
                connector.cache_enabled = True
                
                # Criar request para teste
                request = AccessRequest(
                    request_id="test-req-003",
                    user_context={"user_id": "user789", "groups": ["users"]},
                    resource_context={"resource_id": "doc123", "resource_type": "document", "action": "read"},
                    environment_context={"ip": "10.0.0.1", "device": "desktop"},
                    tenant_id="test-tenant",
                )
                
                # Primeira chamada - deve acionar requisição HTTP
                result1 = await connector.evaluate_access(request)
                
                # Segunda chamada - deve usar cache
                result2 = await connector.evaluate_access(request)
                
                # Verificar resultados
                assert result1.decision == result2.decision
                assert result1.risk_level == result2.risk_level
                
                # Verificar que a requisição HTTP foi feita apenas uma vez
                assert mock_client.request.call_count == 1


if __name__ == "__main__":
    pytest.main()